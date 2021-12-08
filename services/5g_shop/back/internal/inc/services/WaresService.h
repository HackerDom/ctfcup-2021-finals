#ifndef INTERNAL_WARESSERVICE_H
#define INTERNAL_WARESSERVICE_H

#include <memory>
#include <vector>

#include "tools/Result.h"
#include "models/PGConnectionPool.h"
#include "models/Ware.h"

namespace shop {
    class WaresService {
    public:
        explicit WaresService(std::shared_ptr<PGConnectionPool> pgConnectionPool);

        Result<std::string> Create(int sellerId, const std::string &title, const std::string &description, int price);

        Result<std::vector<std::shared_ptr<Ware>>> GetWaresOfUser(int userId);

        Result<std::shared_ptr<Ware>> Get(int id);

    private:
        std::shared_ptr<PGConnectionPool> pgConnectionPool;
    };
}

#endif //INTERNAL_WARESSERVICE_H
