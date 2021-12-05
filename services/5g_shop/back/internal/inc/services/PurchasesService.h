#ifndef INTERNAL_PURCASESSERVICE_H
#define INTERNAL_PURCASESSERVICE_H

#include <memory>

#include "models/PGConnectionPool.h"
#include "models/Purchase.h"
#include "tools/Result.h"

namespace shop {
    class PurchasesService {
    public:
        explicit PurchasesService(std::shared_ptr<PGConnectionPool> pgConnectionPool);

        JustResult Create(int buyerId, int wareId);

    private:
        std::shared_ptr<PGConnectionPool> pgConnectionPool;
    };
}

#endif //INTERNAL_PURCASESSERVICE_H
