#ifndef INTERNAL_PURCASESSERVICE_H
#define INTERNAL_PURCASESSERVICE_H

#include <memory>
#include <vector>
#include <string>

#include "models/PGConnectionPool.h"
#include "models/Purchase.h"
#include "tools/Result.h"

namespace shop {
    class PurchasesService {
    public:
        explicit PurchasesService(std::shared_ptr<PGConnectionPool> pgConnectionPool);

        Result<std::string> Create(int buyerId, int wareId);

        Result<std::vector<std::shared_ptr<Purchase>>> GetOfUser(int userId);

    private:
        std::shared_ptr<PGConnectionPool> pgConnectionPool;
    };
}

#endif //INTERNAL_PURCASESSERVICE_H
