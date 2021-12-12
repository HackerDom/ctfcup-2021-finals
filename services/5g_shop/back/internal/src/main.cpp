#include <iostream>
#include <random>
#include <fstream>
#include <algorithm>

#include <sys/stat.h>
#include <unistd.h>

#include "argparse/CommandLineParser.h"
#include "crow/crow_all.h"
#include "tools/Strings.h"
#include "services/ShopService.h"
#include "tools/SHA256.h"

using namespace shop;
using namespace crow;

typedef App<CookieParser> App;

const char *AuthCookieName = "5GAuth";

response CreateUser(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto json = json::load(req.body);
    if (!json || !json.has("login") || !json.has("password_hash") || !json.has("credit_card_info")) {
        return response(status::BAD_REQUEST);
    }

    auto result = usersService->Create(json["login"].s(), json["password_hash"].s(), json["credit_card_info"].s());

    if (!result.success) {
        return response(status::CONFLICT);
    }

    auto &ctx = app.get_context<CookieParser>(req);

    ctx.set_cookie(AuthCookieName, result.value->authCookie);

    json::wvalue data({
                              {"id",          result.value->id},
                              {"auth_cookie", result.value->authCookie},
                              {"cashback",    result.value->cashback}
                      });

    return {status::CREATED, data};
}

response UserAuth(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto json = json::load(req.body);
    if (!json || !json.has("login") || !json.has("password_hash")) {
        return response(status::BAD_REQUEST);
    }

    auto result = usersService->FindByLoginAndPassword(json["login"].s(), json["password_hash"].s());

    if (!result.success) {
        return response(status::UNAUTHORIZED);
    }

    auto &ctx = app.get_context<CookieParser>(req);
    ctx.set_cookie(AuthCookieName, result.value->authCookie);

    json::wvalue data(
            {
                    {"id",               result.value->id},
                    {"login",            result.value->login},
                    {"created_at",       result.value->createdAt},
                    {"credit_card_info", result.value->creditCardInfo},
                    {"cashback",         result.value->cashback}
            });

    return {status::OK, data};
}

response GetUser(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userResult = usersService->FindByAuthCookie(authCookie);

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

response GetUser(::App &app, const request &req, ShopService &shop, int id) {
    auto usersService = shop.usersService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto authUser = usersService->FindByAuthCookie(authCookie);

    if (!authUser.success) {
        return response(UNAUTHORIZED);
    }

    auto userResult = usersService->GetById(id);

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

response CreateWare(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto waresService = shop.waresService;
    auto json = json::load(req.body);
    if (!json || !json.has("title") || !json.has("description")
        || !json.has("price") || json["price"].i() <= 0 || json["price"].i() > 3 * 100500
        || !json.has("image_id")) {
        return response(status::BAD_REQUEST);
    }

    auto &ctx = app.get_context<CookieParser>(req);
    auto auth = ctx.get_cookie(AuthCookieName);

    if (auth.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto user = usersService->FindByAuthCookie(auth);

    if (!user.success) {
        return response(status::UNAUTHORIZED);
    }

    auto wareCreated = waresService->Create(
            user.value->id, json["title"].s(),
            json["description"].s(), static_cast<int>(json["price"].i()),
            static_cast<int>(json["image_id"].i()));

    if (!wareCreated.success) {
        CROW_LOG_ERROR << "ware creation failed: " << wareCreated.message;

        return response(status::INTERNAL_SERVER_ERROR);
    }

    json::wvalue data({
                              {"id", wareCreated.value}
                      });

    return {status::CREATED, data};
}

response GetWare(::App &app, const request &req, ShopService &shop, int id) {
    auto usersService = shop.usersService;
    auto waresService = shop.waresService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userResult = usersService->FindByAuthCookie(authCookie);

    if (!userResult.success) {
        return response(status::UNAUTHORIZED);
    }

    auto wareFindResult = waresService->Get(id);

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
                    {"price",       ware->price},
                    {"image_id",    ware->imageId}
            });

    if (ware->sellerId == userResult.value->id) {
        data["service_fee"] = ware->serviceFee;
    }

    return {status::OK, data};
}

response GetWaresOfUser(::App &app, const request &req, ShopService &shop, int userId = -1) {
    auto usersService = shop.usersService;
    auto waresService = shop.waresService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userResult = usersService->FindByAuthCookie(authCookie);

    if (!userResult.success) {
        return response(status::UNAUTHORIZED);
    }

    auto waresResult = waresService->GetWaresOfUser(userId < 0 ? userResult.value->id : userId);

    std::vector<json::wvalue> wareIds;

    for (const auto &ware: waresResult.value) {
        wareIds.emplace_back(json::wvalue(ware->id));
    }

    json::wvalue data({{"ids", wareIds}});

    return {status::OK, data};
}

response ListUsers(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService->FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    if (req.url_params.get("page_num") == nullptr || req.url_params.get("page_size") == nullptr) {
        return response(status::BAD_REQUEST);
    }

    auto pageNum = std::stoi(std::string("0") + req.url_params.get("page_num"));
    auto pageSize = std::stoi(std::string("0") + req.url_params.get("page_size"));

    auto list = usersService->List(pageNum, pageSize);

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

response MakePurchase(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto purchasesService = shop.purchasesService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService->FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    if (req.url_params.get("ware_id") == nullptr) {
        return response(status::BAD_REQUEST);
    }

    auto wareId = std::stoi(std::string("0") + req.url_params.get("ware_id"));
    auto result = purchasesService->Create(userAuth.value->id, wareId);

    if (!result.success) {
        return response(status::NOT_FOUND);
    }

    json::wvalue data(
            {
                    {"id", result.value}
            }
    );

    return {status::CREATED, data};
}

response GetPurchasesOfUser(::App &app, const request &req, ShopService &shop, int userId = -1) {
    auto usersService = shop.usersService;
    auto purchasesService = shop.purchasesService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService->FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    auto purchases = purchasesService->GetOfUser(userId < 0 ? userAuth.value->id : userId);

    if (!purchases.success) {
        return response(status::INTERNAL_SERVER_ERROR);
    }

    std::vector<json::wvalue> jsoned;

    for (const auto &purchase: purchases.value) {
        json::wvalue data(
                {
                        {"id",      purchase->id},
                        {"ware_id", purchase->wareId}
                });

        jsoned.push_back(data);
    }

    json::wvalue data({{"purchases", jsoned}});

    return {status::OK, data};
}

bool FileExists(const std::string &name) {
    struct stat buffer{};
    return (stat(name.c_str(), &buffer) == 0);
}

std::random_device dev;
std::mt19937 rng(dev());
std::uniform_int_distribution<std::mt19937::result_type> filesSuffixDist(0);

std::string GenerateFileName(const std::string &prefix) {
    std::string currentName = prefix;

    while (FileExists(currentName)) {
        currentName = prefix + std::to_string(filesSuffixDist(rng));
    }

    return currentName;
}

response UploadImage(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto imagesService = shop.imagesService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService->FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    crow::multipart::message msg(req);

    if (msg.parts.size() != 1) {
        return response(status::INTERNAL_SERVER_ERROR);
    }

    auto &part = msg.parts[0];
    std::string filename;

    for (auto &header: part.headers) {
        if (header.value.first == "Content-Type") {
            if (!header.value.second.starts_with("image")) {
                return response(status::BAD_REQUEST);
            }
        }
        if (header.value.first == "Content-Disposition") {
            filename = header.params["filename"];
        }
    }

    Trim(filename);
    std::replace(filename.begin(), filename.end(), '/', '_');
    std::replace(filename.begin(), filename.end(), ' ', '_');

    if (filename.empty() || part.body.size() > 5 * 1024 * 1024) {
        return response(status::BAD_REQUEST, "invalid filename or file is too large");
    }

    auto filecontent = part.body.c_str();

    SHA256 sha;
    sha.update(reinterpret_cast<const uint8_t *>(filecontent), part.body.size());
    auto digest = sha.digest();
    std::string sha256 = SHA256::toString(digest.get());

    Result<std::shared_ptr<Image>> bySha256 = imagesService->FindBySha256(sha256);

    if (bySha256.success) {
        Result<std::shared_ptr<Image>> image = imagesService->Create(userAuth.value->id, bySha256.value->filename);

        if (!image.success) {
            return {status::INTERNAL_SERVER_ERROR, image.message};
        }

        json::wvalue data(
                {
                        {"path", bySha256.value->filename},
                        {"id",   image.value->id}
                });

        return {status::CREATED, data};
    }

    filename = GenerateFileName("/images/" + filename);
    auto path = "/api/images/get/" + filename.substr(8);

    std::ofstream file(filename, std::ios_base::out | std::ios_base::binary);

    if (file.bad()) {
        return response(status::INTERNAL_SERVER_ERROR);
    }

    file.write(filecontent, static_cast<long>(part.body.size()));

    file.close();

    Result<std::shared_ptr<Image>> image = imagesService->Create(userAuth.value->id, path);

    if (!image.success) {
        return {status::INTERNAL_SERVER_ERROR, image.message};
    }

    json::wvalue data(
            {
                    {"path", path},
                    {"id",   image.value->id}
            });

    return {status::CREATED, data};
}

response GetImage(::App &app, const request &req, ShopService &shop, int id) {
    auto usersService = shop.usersService;
    auto imagesService = shop.imagesService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService->FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    auto image = imagesService->Get(id);
    if (!image.success) {
        return response(status::NOT_FOUND);
    }

    json::wvalue data(
            {
                    {"id",       image.value->id},
                    {"owner_id", image.value->ownerId},
                    {"path",     image.value->filename}
            });

    return {status::OK, data};
}

response GetMyImages(::App &app, const request &req, ShopService &shop) {
    auto usersService = shop.usersService;
    auto imagesService = shop.imagesService;
    auto &ctx = app.get_context<CookieParser>(req);
    auto authCookie = ctx.get_cookie(AuthCookieName);

    if (authCookie.empty()) {
        return response(status::UNAUTHORIZED);
    }

    auto userAuth = usersService->FindByAuthCookie(authCookie);

    if (!userAuth.success) {
        return response(status::UNAUTHORIZED);
    }

    auto images = imagesService->ListOfUser(userAuth.value->id);
    if (!images.success) {
        return {status::INTERNAL_SERVER_ERROR, images.message};
    }

    std::vector<json::wvalue> jsoned;

    for (const auto &image: images.value) {
        json::wvalue data(
                {
                        {"id",   image->id},
                        {"path", image->filename}
                });

        jsoned.push_back(data);
    }

    json::wvalue data({{"images", jsoned}});

    return {status::OK, data};
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
        auto usersService = std::make_shared<UsersService>(pgConnectionPool);
        auto waresService = std::make_shared<WaresService>(pgConnectionPool);
        auto purchasesService = std::make_shared<PurchasesService>(pgConnectionPool);
        auto imagesServices = std::make_shared<ImagesService>(pgConnectionPool);
        ShopService shopService(usersService, imagesServices, waresService, purchasesService);

        CROW_ROUTE(app, "/api/users").methods(HTTPMethod::Post, HTTPMethod::Get)(
                [&shopService, &app](const request &req) {
                    if (req.method == HTTPMethod::Post) {
                        return CreateUser(app, req, shopService);
                    }

                    return GetUser(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/users/<int>").methods(HTTPMethod::Get)(
                [&app, &shopService](const request &req, int userId) {
                    return GetUser(app, req, shopService, userId);
                });

        CROW_ROUTE(app, "/api/users/list").methods(HTTPMethod::Get)(
                [&app, &shopService](const request &req) {
                    return ListUsers(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/users/auth").methods(HTTPMethod::Put)(
                [&shopService, &app](const request &req) {
                    return UserAuth(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/wares").methods(HTTPMethod::Post)(
                [&shopService, &app](const request &req) {
                    return CreateWare(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/wares/<int>").methods(HTTPMethod::Get)(
                [&app, &shopService](const request &req, int wareId) {
                    return GetWare(app, req, shopService, wareId);
                });

        CROW_ROUTE(app, "/api/wares/my").methods(HTTPMethod::Get)(
                [&app, &shopService](const request &req) {
                    return GetWaresOfUser(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/wares/of_user/<int>").methods(HTTPMethod::Get)(
                [&app, &shopService](const request &req, int userId) {
                    return GetWaresOfUser(app, req, shopService, userId);
                });

        CROW_ROUTE(app, "/api/purchases").methods(HTTPMethod::Post)(
                [&app, &shopService](const request &req) {
                    return MakePurchase(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/purchases/my").methods(HTTPMethod::Get)(
                [&app, &shopService](const request &req) {
                    return GetPurchasesOfUser(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/images").methods(HTTPMethod::Post)(
                [&app, &shopService](const request &req) {
                    return UploadImage(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/images/my").methods(crow::HTTPMethod::Get)(
                [&app, &shopService](const request &req) {
                    return GetMyImages(app, req, shopService);
                });

        CROW_ROUTE(app, "/api/images/<int>").methods(crow::HTTPMethod::Get)(
                [&app, &shopService](const request &req, int id) {
                    return GetImage(app, req, shopService, id);
                });

        CROW_ROUTE(app, "/api/images/get/<string>")(
                [&shopService](const request &req, response &res, std::string) {
                    auto imagesService = shopService.imagesService;
                    auto file = imagesService->FindAnyWithFilename(req.url);

                    if (!file.success) {
                        res.code = status::NOT_FOUND;
                        res.end();
                    }

                    res.set_static_file_info("/images/" + file.value->filename.substr(16));
                    res.end();
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
