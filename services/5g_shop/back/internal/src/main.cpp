#include "argparse/CommandLineParser.h"
#include "crow/crow_all.h"
#include "tools/Strings.h"
#include "models/PGConnectionPool.h"
#include "services/UsersService.h"
#include "services/WaresService.h"
#include "services/PurchasesService.h"

using namespace shop;
using namespace crow;

typedef App<CookieParser> App;

const char *AuthCookieName = "5GAuth";

response CreateUser(::App &app, const request &req, UsersService &usersService) {
    auto json = json::load(req.body);
    if (!json || !json.has("login") || !json.has("password_hash") || !json.has("credit_card_info")) {
        return response(status::BAD_REQUEST);
    }

    auto result = usersService.Create(json["login"].s(), json["password_hash"].s(), json["credit_card_info"].s());

    if (!result.success) {
        return response(status::CONFLICT);
    }

    auto &ctx = app.get_context<CookieParser>(req);

    ctx.set_cookie(AuthCookieName, result.value);

    return response(status::CREATED);
}

response UserAuth(::App &app, const request &req, UsersService &usersService) {
    auto json = json::load(req.body);
    if (!json || !json.has("login") || !json.has("password_hash")) {
        return response(status::BAD_REQUEST);
    }

    auto result = usersService.FindByLoginAndPassword(json["login"].s(), json["password_hash"].s());

    if (!result.success) {
        return response(status::UNAUTHORIZED);
    }

    auto &ctx = app.get_context<CookieParser>(req);
    ctx.set_cookie(AuthCookieName, result.value->authCookie);

    return response(status::OK);
}

response GetUser(::App &app, const request &req, UsersService &usersService) {
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userResult = usersService.FindByAuthCookie(authCookie);

    if (!userResult.success) {
        return response(status::UNAUTHORIZED);
    }

    json::wvalue data(
            {
                    {"id",               userResult.value->id},
                    {"login",            userResult.value->login},
                    {"created_at",       userResult.value->createdAt},
                    {"credit_card_info", userResult.value->creditCardInfo},
                    {"cashback",         userResult.value->cashback}
            });

    return {status::OK, data};
}

response GetUser(::App &app, const request &req, UsersService &usersService, int id) {
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto authUser = usersService.FindByAuthCookie(authCookie);

    if (!authUser.success) {
        return response(UNAUTHORIZED);
    }

    auto userResult = usersService.GetById(id);

    if (!userResult.success) {
        return response(status::NOT_FOUND);
    }

    json::wvalue data(
            {
                    {"id",         userResult.value->id},
                    {"login",      userResult.value->login},
                    {"created_at", userResult.value->createdAt}
            });

    if (id == authUser.value->id) {
        data["cashback"] = userResult.value->cashback;
        data["credit_card_info"] = userResult.value->creditCardInfo;
    }

    return {status::OK, data};
}

response CreateWare(::App &app, const request &req, WaresService &waresService, UsersService &usersService) {
    auto json = json::load(req.body);
    if (!json || !json.has("title") || !json.has("description") || !json.has("price") || json["price"].i() <= 0 ||
        json["price"].i() > 100500) {
        return response(status::BAD_REQUEST);
    }

    auto &ctx = app.get_context<CookieParser>(req);
    auto auth = ctx.get_cookie(AuthCookieName);

    if (auth.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto user = usersService.FindByAuthCookie(auth);

    if (!user.success) {
        return response(status::UNAUTHORIZED);
    }

    auto wareCreated = waresService.Create(
            user.value->id, json["title"].s(), json["description"].s(), static_cast<int>(json["price"].i()));

    if (!wareCreated.success) {
        CROW_LOG_ERROR << "ware creation failed: " << wareCreated.message;

        return response(status::INTERNAL_SERVER_ERROR);
    }

    return response(status::CREATED);
}

response
GetWare(::App &app, const request &req, WaresService &waresService, UsersService &usersService, int id) {
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userResult = usersService.FindByAuthCookie(authCookie);

    if (!userResult.success) {
        return response(status::UNAUTHORIZED);
    }

    auto wareFindResult = waresService.Get(id);

    if (!wareFindResult.success) {
        return response(status::NOT_FOUND);
    }

    auto ware = wareFindResult.value;

    json::wvalue data(
            {
                    {"id",          ware->id},
                    {"seller_id",   ware->sellerId},
                    {"title",       ware->title},
                    {"description", ware->description},
                    {"price",       ware->price}
            });

    if (ware->sellerId == userResult.value->id) {
        data["service_fee"] = ware->serviceFee;
    }

    return {status::OK, data};
}

response
GetWaresOfUser(::App &app, const request &req, WaresService &waresService, UsersService &usersService,
               int userId = -1) {
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userResult = usersService.FindByAuthCookie(authCookie);

    if (!userResult.success) {
        return response(status::UNAUTHORIZED);
    }

    auto waresResult = waresService.GetWaresOfUser(userId < 0 ? userResult.value->id : userId);

    std::vector<json::wvalue> wareIds;

    for (const auto &ware: waresResult.value) {
        wareIds.emplace_back(json::wvalue(ware->id));
    }

    json::wvalue data({{"ids", wareIds}});

    return {status::OK, data};
}

response ListUsers(::App &app, const request &req, UsersService &usersService) {
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService.FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    if (req.url_params.get("page_num") == nullptr || req.url_params.get("page_size") == nullptr) {
        return response(status::BAD_REQUEST);
    }

    auto pageNum = std::stoi(std::string("0") + req.url_params.get("page_num"));
    auto pageSize = std::stoi(std::string("0") + req.url_params.get("page_size"));

    auto list = usersService.List(pageNum, pageSize);

    if (!list.success) {
        return response(status::INTERNAL_SERVER_ERROR);
    }

    std::vector<json::wvalue> jsoned;

    for (const auto &user: list.value) {
        json::wvalue data(
                {
                        {"id",         user->id},
                        {"login",      user->login},
                        {"created_at", user->createdAt}
                });

        if (user->id == userAuth.value->id) {
            data["cashback"] = user->cashback;
            data["credit_card_info"] = user->creditCardInfo;
        }

        jsoned.push_back(data);
    }

    json::wvalue data({{"users", jsoned}});

    return {status::OK, data};
}

response MakePurchase(::App &app, const request &req, UsersService &usersService, PurchasesService &purchasesService) {
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService.FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    if (req.url_params.get("ware_id") == nullptr) {
        return response(status::BAD_REQUEST);
    }

    auto wareId = std::stoi(std::string("0") + req.url_params.get("ware_id"));
    auto result = purchasesService.Create(userAuth.value->id, wareId);

    if (!result.success) {
        return response(status::NOT_FOUND);
    }

    return response(status::CREATED);
}

int main(int argc, char **argv, char **env) {
    try {
        auto args = CommandLineParser::Parse(argc, argv);

        ::App app;

        PGConnectionConfig pgConfig = {
                GetEnv("POSTGRES_HOST"),
                GetIntEnv("POSTGRES_PORT"),
                GetEnv("POSTGRES_DB"),
                GetEnv("POSTGRES_USER"),
                GetEnv("POSTGRES_PASSWORD")
        };
        auto pgConnectionPool = std::make_shared<PGConnectionPool>(pgConfig, 10);
        UsersService usersService(pgConnectionPool);
        WaresService waresService(pgConnectionPool);
        PurchasesService purchasesService(pgConnectionPool);

        CROW_ROUTE(app, "/api/users").methods(HTTPMethod::Post, HTTPMethod::Get)(
                [&usersService, &app](const request &req) {
                    if (req.method == HTTPMethod::Post) {
                        return CreateUser(app, req, usersService);
                    }

                    return GetUser(app, req, usersService);
                });

        CROW_ROUTE(app, "/api/users/<int>").methods(HTTPMethod::Get)(
                [&app, &usersService](const request &req, int userId) {
                    return GetUser(app, req, usersService, userId);
                });

        CROW_ROUTE(app, "/api/users/list").methods(HTTPMethod::Get)(
                [&app, &usersService](const request &req) {
                    return ListUsers(app, req, usersService);
                });

        CROW_ROUTE(app, "/api/users/auth").methods(HTTPMethod::Put)(
                [&usersService, &app](const request &req) {
                    return UserAuth(app, req, usersService);
                });

        CROW_ROUTE(app, "/api/wares").methods(HTTPMethod::Post)(
                [&usersService, &waresService, &app](const request &req) {
                    return CreateWare(app, req, waresService, usersService);
                });

        CROW_ROUTE(app, "/api/wares/<int>").methods(HTTPMethod::Get)(
                [&app, &usersService, &waresService](const request &req, int wareId) {
                    return GetWare(app, req, waresService, usersService, wareId);
                });

        CROW_ROUTE(app, "/api/wares/my").methods(HTTPMethod::Get)(
                [&app, &usersService, &waresService](const request &req) {
                    return GetWaresOfUser(app, req, waresService, usersService);
                });

        CROW_ROUTE(app, "/api/wares/of_user/<int>").methods(HTTPMethod::Get)(
                [&app, &usersService, &waresService](const request &req, int userId) {
                    return GetWaresOfUser(app, req, waresService, usersService, userId);
                });

        CROW_ROUTE(app, "/api/purchases").methods(crow::HTTPMethod::Post)(
                [&app, &usersService, &purchasesService](const request &req) {
                    return MakePurchase(app, req, usersService, purchasesService);
                });

        app
                .server_name("5G shop backend")
                .bindaddr(args.address)
                .port(args.port)
                .concurrency(std::thread::hardware_concurrency() - 1)
                .run();

    } catch (std::exception &e) {
        CROW_LOG_CRITICAL << "Unhandled exception: " << e.what();

        return EXIT_FAILURE;
    } catch (...) {
        CROW_LOG_CRITICAL << "Unknown unhandled exception occurred";

        return EXIT_FAILURE;
    }
}
