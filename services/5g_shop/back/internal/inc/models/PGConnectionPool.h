#ifndef INTERNAL_PGPOOL_H
#define INTERNAL_PGPOOL_H

#include <memory>
#include <mutex>
#include <queue>
#include <condition_variable>

#include "models/PGConnection.h"

namespace shop {
    class PGConnectionPool;

    class PGConnectionGuard {
    public:
        PGConnectionGuard(std::shared_ptr<PGConnection> connection, PGConnectionPool *owner);

        PGConnectionGuard(const PGConnectionGuard &other) = delete;

        PGConnectionGuard(PGConnectionGuard &&other) = default;

        PGConnectionGuard &operator=(const PGConnectionGuard &other) = delete;

        PGConnectionGuard &operator=(PGConnectionGuard &&other) = delete;

        ~PGConnectionGuard();

        const std::shared_ptr<PGConnection> connection;
    private:
        PGConnectionPool *owner;
    };

    class PGConnectionPool {
    public:
        PGConnectionPool(const PGConnectionConfig &config, int connectionCount);

        std::shared_ptr<PGConnection> Get();

        void Free(const std::shared_ptr<PGConnection> &connection);

        PGConnectionGuard Guarded();

    private:
        std::queue<std::shared_ptr<PGConnection>> connections;
        std::condition_variable freeConnectionNotifier;
        std::mutex mutex;
    };
}

#endif //INTERNAL_PGPOOL_H
