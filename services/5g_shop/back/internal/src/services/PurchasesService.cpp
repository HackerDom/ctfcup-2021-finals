#include <utility>

#include "crow/crow_all.h"

#include "services/PurchasesService.h"
#include "tools/Defer.h"
#include "tools/Strings.h"

using namespace shop;

PurchasesService::PurchasesService(std::shared_ptr<PGConnectionPool> pgConnectionPool)
        : pgConnectionPool(std::move(pgConnectionPool)) {
}

JustResult PurchasesService::Create(int buyerId, int wareId) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = format(
            HiddenStr("insert into purchases values (default, %d, %d);"),
            wareId,
            buyerId
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_COMMAND_OK) {
        auto msg = std::string(PQerrorMessage(conn));

        CROW_LOG_ERROR << "error creating new purchase: " << msg;

        return JustResult::ofError(msg);
    }

    return JustResult::ofSuccess();
}