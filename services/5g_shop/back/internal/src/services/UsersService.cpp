#include <utility>

#include <libpq-fe.h>

#include "crow/crow_all.h"

#include "services/UsersService.h"
#include "tools/Strings.h"
#include "tools/Defer.h"

#include "tools/UUID.h"
#include "money/Calc.h"

using namespace shop;

Result<std::shared_ptr<User>>
UsersService::Create(const std::string &login, const std::string &passwordHash, const std::string &creditCardInfo) {
    Defer defer;
    std::string authCookie = UUID4();
    auto cashback = GetCashback(login);

    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = connectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto escapedLogin = Escape(conn, login);
    auto passwordEscaped = Escape(conn, passwordHash);
    auto authCookieEscaped = Escape(conn, authCookie);
    auto cardInfoEscaped = Escape(conn, creditCardInfo);

    auto query = Format(
            HiddenStr("insert into users values (default, %s, default, %s, %s, %s, %d) returning *;"),
            escapedLogin.c_str(),
            passwordEscaped.c_str(),
            authCookieEscaped.c_str(),
            cardInfoEscaped.c_str(),
            cashback
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "creation of user '" << login << "' failed with: " << msg;

        return Result<std::shared_ptr<User>>::ofError("possible login conflict");
    }

    return Result<std::shared_ptr<User>>::ofSuccess(User::ReadFromPGResult(result));
}

Result<std::shared_ptr<User>> UsersService::FindByAuthCookie(const std::string &authCookie) {
    Defer defer;
    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = connectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto escapedCookie = Escape(conn, authCookie);

    auto query = Format(
            HiddenStr("select * from users where auth_cookie=%s;"),
            escapedCookie.c_str()
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        return Result<std::shared_ptr<User>>::ofError("user with that cookie not found");
    }

    return Result<std::shared_ptr<User>>::ofSuccess(User::ReadFromPGResult(result));
}

Result<std::shared_ptr<User>>
UsersService::FindByLoginAndPassword(const std::string &login, const std::string &passwordHash) {
    Defer defer;
    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = connectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

//    auto escapedLogin = Escape(conn, login);
    auto escapedPasswordHash = Escape(conn, passwordHash);

    auto query = Format(
            HiddenStr("select * from users where login='%s' and password_hash=%s;"),
            login.c_str(),
            escapedPasswordHash.c_str()
    );

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        auto msg = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << msg;

        return Result<std::shared_ptr<User>>::ofError("invalid login or password");
    }

    auto user = User::ReadFromPGResult(result);

    if (user->passwordHash != passwordHash || user->login != login) {
        return Result<std::shared_ptr<User>>::ofError("invalid login or password");
    }

    return Result<std::shared_ptr<User>>::ofSuccess(user);
}

Result<std::shared_ptr<User>> UsersService::GetById(int id) {
    Defer defer;
    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    auto guard = connectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(HiddenStr("select * from users where id=%d"), id);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK || PQntuples(result) != 1) {
        return Result<std::shared_ptr<User>>::ofError("invalid login or password");
    }

    return Result<std::shared_ptr<User>>::ofSuccess(User::ReadFromPGResult(result));
}

Result<std::vector<std::shared_ptr<User>>> UsersService::List(int pageNum, int pageSize) {
    Defer defer;
    PGresult *result = nullptr;
    defer([&result] {
        PQclear(result);
    });

    if (pageSize <= 0) {
        pageSize = 10;
    }

    pageSize = std::min(pageSize, 100);

    if (pageNum < 0) {
        pageNum = 0;
    }

    auto guard = connectionPool->Guarded();
    auto conn = guard.connection->Connection().get();

    auto query = Format(HiddenStr("select * from users limit %d offset %d;"), pageSize, pageNum * pageSize);

    result = PQexec(conn, query.c_str());

    if (PQresultStatus(result) != PGRES_TUPLES_OK) {
        auto err = std::string(PQresultErrorMessage(result));

        CROW_LOG_ERROR << "users listing failed: " << err;

        return Result<std::vector<std::shared_ptr<User>>>::ofError();
    }

    std::vector<std::shared_ptr<User>> users;

    int n = PQntuples(result);

    for (int i = 0; i < n; ++i) {
        auto user = User::ReadFromPGResult(result, i);

        users.push_back(user);
    }

    return Result<std::vector<std::shared_ptr<User>>>::ofSuccess(users);
}

UsersService::UsersService(std::shared_ptr<PGConnectionPool> connectionPool)
        : connectionPool(std::move(connectionPool)) {
}
