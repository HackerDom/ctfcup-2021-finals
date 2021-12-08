#include "models/PGConnectionPool.h"

#include <utility>

using namespace shop;

PGConnectionGuard::PGConnectionGuard(std::shared_ptr<PGConnection> connection, PGConnectionPool *owner)
        : connection(std::move(connection)),
          owner(owner) {
}

PGConnectionGuard::~PGConnectionGuard() {
    owner->Free(connection);
}

PGConnectionPool::PGConnectionPool(const PGConnectionConfig &config, int connectionCount) {
    std::scoped_lock<std::mutex> lock(mutex);

    connectionCount = std::max(1, connectionCount);

    for (auto i = 0; i < connectionCount; ++i) {
        connections.emplace(std::make_shared<PGConnection>(config));
    }
}

std::shared_ptr<PGConnection> PGConnectionPool::Get() {
    std::unique_lock<std::mutex> lock(mutex);

    freeConnectionNotifier.wait(lock, [this] { return !connections.empty(); });

    auto connection = connections.front();
    connections.pop();

    return connection;
}

void PGConnectionPool::Free(const std::shared_ptr<PGConnection>& connection) {
    std::unique_lock<std::mutex> lock(mutex);

    connections.push(connection);

    lock.unlock();

    freeConnectionNotifier.notify_one();
}

PGConnectionGuard PGConnectionPool::Guarded() {
    return {Get(), this};
}
