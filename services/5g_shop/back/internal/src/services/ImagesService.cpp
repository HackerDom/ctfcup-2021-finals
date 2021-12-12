#include "crow/crow_all.h"

#include "services/ImagesService.h"
#include "tools/Defer.h"
#include "tools/Strings.h"

using namespace shop;

ImagesService::ImagesService(std::shared_ptr<PGConnectionPool> pgConnectionPool)
        : pgConnectionPool(std::move(pgConnectionPool)) {
}

Result<std::shared_ptr<Image>> ImagesService::Get(int id) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(HiddenStr("select * from images where id=%d;"), id);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "image registration failed with " << msg;

        return Result<std::shared_ptr<Image>>::ofError(msg);
    }

    return Result<std::shared_ptr<Image>>::ofSuccess(Image::ReadFromPGResult(result));
}


Result<std::shared_ptr<Image>> ImagesService::FindBySha256(const std::string &sha256) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(HiddenStr("select * from images where sha256='%s' limit 1;"), sha256.c_str());

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        return Result<std::shared_ptr<Image>>::ofError();
    }

    return Result<std::shared_ptr<Image>>::ofSuccess(Image::ReadFromPGResult(result));
}


Result<std::shared_ptr<Image>> ImagesService::Create(int ownerId, const std::string &filename) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(
            HiddenStr(
                    "insert into images values (default, %d, %s, encode(sha256(pg_read_binary_file(%s, 0, 10000000)), 'hex')) returning *;"),
            ownerId,
            Escape(conn, filename).c_str(),
            Escape(conn, ("./" + filename.substr(16))).c_str()
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "image registration failed with " << msg;

        return Result<std::shared_ptr<Image>>::ofError(msg);
    }

    return Result<std::shared_ptr<Image>>::ofSuccess(Image::ReadFromPGResult(result));
}

Result<std::vector<std::shared_ptr<Image>>> ImagesService::ListOfUser(int ownerId) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(HiddenStr("select * from images where owner_id=%d;"), ownerId);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "image list for user failed with " << msg;

        return Result<std::vector<std::shared_ptr<Image>>>::ofError(msg);
    }

    std::vector<std::shared_ptr<Image>> wares;

    int n = PQntuples(result);

    for (int i = 0; i < n; ++i) {
        auto ware = Image::ReadFromPGResult(result, i);

        wares.push_back(ware);
    }

    return Result<std::vector<std::shared_ptr<Image>>>::ofSuccess(wares);
}

Result<std::shared_ptr<Image>> ImagesService::FindAnyWithFilename(const std::string &filename) {
    Defer defer;

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = pgConnectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(HiddenStr("select * from images where filename=%s limit 1;"), Escape(conn, filename).c_str());

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        return Result<std::shared_ptr<Image>>::ofError();
    }

    return Result<std::shared_ptr<Image>>::ofSuccess(Image::ReadFromPGResult(result));
}
