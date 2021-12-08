#include <utility>

#include "crow/crow_all.h"

#include "services/PurchasesService.h"
#include "tools/Defer.h"
#include "tools/Strings.h"

using namespace shop;

PurchasesService::PurchasesService(std::shared_ptr<PGConnectionPool> pgConnectionPool)
        : pgConnectionPool(std::move(pgConnectionPool)) {
}

Result<std::string> PurchasesService::Create(int buyerId, int wareId) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = format(
            HiddenStr("insert into purchases values (default, %d, %d) returning id;"),
            wareId,
            buyerId
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "error creating new purchase: " << msg;

        return Result<std::string>::ofError(msg);
    }

    return Result<std::string>::ofSuccess(std::string(PQgetvalue(result, 0, 0)));
}

Result<std::vector<std::shared_ptr<Purchase>>> PurchasesService::GetOfUser(int userId) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = format(HiddenStr("select * from purchases where buyer_id=%d;"), userId);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "error selecting purchases of user: " << msg;

        return Result<std::vector<std::shared_ptr<Purchase>>>::ofError(msg);
    }

    std::vector<std::shared_ptr<Purchase>> purchases;

    int n = PQntuples(result);

    for (int i = 0; i < n; ++i) {
        auto purchase = Purchase::ReadFromPGResult(result, i);

        purchases.push_back(purchase);
    }

    return Result<std::vector<std::shared_ptr<Purchase>>>::ofSuccess(purchases);
}
