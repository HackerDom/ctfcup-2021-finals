#include <utility>

#include "crow/crow_all.h"

#include "services/WaresService.h"
#include "tools/Defer.h"
#include "tools/Strings.h"
#include "money/Calc.h"

using namespace shop;

WaresService::WaresService(std::shared_ptr<PGConnectionPool> pgConnectionPool)
        : pgConnectionPool(std::move(pgConnectionPool)) {
}

Result<std::string> WaresService::Create(int sellerId, const std::string &title, const std::string &description, int price) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    int serviceFee = GetServiceFee(price, title, description);

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = format(
            HiddenStr("insert into wares values (default, %d, '%s', '%s', %d, %d) returning id;"),
            sellerId,
            title.c_str(),
            description.c_str(),
            price,
            serviceFee
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        return Result<std::string>::ofError(PQerrorMessage(conn));
    }

    return Result<std::string>::ofSuccess(std::string(PQgetvalue(result, 0, 0)));
}

Result<std::shared_ptr<Ware>> WaresService::Get(int id) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = format(HiddenStr("select * from wares where id=%d;"), id);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        return Result<std::shared_ptr<Ware>>::ofError();
    }

    return Result<std::shared_ptr<Ware>>::ofSuccess(Ware::ReadFromPGResult(result));
}

Result<std::vector<std::shared_ptr<Ware>>> WaresService::GetWaresOfUser(int userId) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = format(HiddenStr("select * from wares where seller_id=%d;"), userId);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK) {
        std::string err = PQerrorMessage(conn);

        CROW_LOG_ERROR << "select wares of user failed: " << err;

        return Result<std::vector<std::shared_ptr<Ware>>>::ofError(err);
    }

    std::vector<std::shared_ptr<Ware>> wares;

    int n = PQntuples(result);

    for (int i = 0; i < n; ++i) {
        auto ware = Ware::ReadFromPGResult(result, i);

        wares.push_back(ware);
    }

    return Result<std::vector<std::shared_ptr<Ware>>>::ofSuccess(wares);
}
