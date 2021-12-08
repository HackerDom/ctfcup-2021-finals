#ifndef INTERNAL_PGCONNECTION_H
#define INTERNAL_PGCONNECTION_H

#include <string>
#include <memory>

#include <libpq-fe.h>

namespace shop {
    struct PGConnectionConfig {
        const std::string host;
        const int port;
        const std::string dbName;
        const std::string user;
        const std::string password;
    };

    class PGConnection {
    public:
        explicit PGConnection(const PGConnectionConfig &config);

        PGConnection(const PGConnection &other) = delete;

        PGConnection(PGConnection &&other) = delete;

        PGConnection &operator=(const PGConnection &other) = delete;

        PGConnection &operator=(PGConnection &&other) = delete;

        [[nodiscard]] std::shared_ptr<PGconn> Connection() const;

    private:
        std::shared_ptr<PGconn> connection;
    };
}

#endif //INTERNAL_PGCONNECTION_H
