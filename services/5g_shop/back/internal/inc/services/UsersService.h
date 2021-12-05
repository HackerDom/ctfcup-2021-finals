#ifndef INTERNAL_USERSSERVICE_H
#define INTERNAL_USERSSERVICE_H

#include <string>

#include "tools/Result.h"
#include "models/User.h"
#include "models/PGConnectionPool.h"

namespace shop {
    class UsersService {
    public:
        explicit UsersService(std::shared_ptr<PGConnectionPool> connectionPool);

        Result<std::string> Create(const std::string &login, const std::string &passwordHash, const std::string &creditCardInfo);

        Result<std::shared_ptr<User>> FindByAuthCookie(const std::string &authCookie);

        Result<std::shared_ptr<User>> FindByLoginAndPassword(const std::string &login, const std::string &passwordHash);

        Result<std::shared_ptr<User>> GetById(int id);

        Result<std::vector<std::shared_ptr<User>>> List(int pageNum, int pageSize);

    private:
        std::shared_ptr<PGConnectionPool> connectionPool;
    };
}

#endif //INTERNAL_USERSSERVICE_H
