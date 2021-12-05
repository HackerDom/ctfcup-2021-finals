#ifndef INTERNAL_USER_H
#define INTERNAL_USER_H

#include <memory>

#include <libpq-fe.h>

namespace shop {
    class User {
    public:
        User(int id, std::string login, std::string passwordHash, std::string authCookie, std::string createdAt,
             std::string creditCardInfo, int cashback)
                : id(id),
                  login(std::move(login)),
                  passwordHash(std::move(passwordHash)),
                  authCookie(std::move(authCookie)),
                  createdAt(std::move(createdAt)),
                  creditCardInfo(std::move(creditCardInfo)),
                  cashback(cashback) {
        }

        const int id;
        const std::string login;
        const std::string passwordHash;
        const std::string authCookie;
        const std::string createdAt;
        const std::string creditCardInfo;
        const int cashback;

        static std::shared_ptr<User> ReadFromPGResult(PGresult *result, int rowNum = 0) {
            auto idCol = PQfnumber(result, "id");
            auto createdAtCol = PQfnumber(result, "created_at");
            auto loginCol = PQfnumber(result, "login");
            auto passwordHashCol = PQfnumber(result, "password_hash");
            auto authCookieCol = PQfnumber(result, "auth_cookie");
            auto creditCardInfoCol = PQfnumber(result, "credit_card_info");
            auto cashbackCol = PQfnumber(result, "cashback");

            if (createdAtCol == -1 || loginCol == -1 || passwordHashCol == -1 || authCookieCol == -1) {
                throw std::runtime_error("unexpected pg result to read user");
            }

            return std::make_shared<User>(
                    std::stoi(std::string(PQgetvalue(result, rowNum, idCol))),
                    std::string(PQgetvalue(result, rowNum, loginCol)),
                    std::string(PQgetvalue(result, rowNum, passwordHashCol)),
                    std::string(PQgetvalue(result, rowNum, authCookieCol)),
                    std::string(PQgetvalue(result, rowNum, createdAtCol)),
                    std::string(PQgetvalue(result, rowNum, creditCardInfoCol)),
                    std::stoi(std::string(PQgetvalue(result, rowNum, cashbackCol)))
            );
        }
    };
}

#endif //INTERNAL_USER_H
