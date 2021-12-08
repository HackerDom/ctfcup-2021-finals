#include "models/PGConnection.h"

using namespace shop;

PGConnection::PGConnection(const PGConnectionConfig &config) {
    pg_conn *pgConn = PQsetdbLogin(config.host.c_str(),
                                   std::to_string(config.port).c_str(),
                                   nullptr,
                                   nullptr,
                                   config.dbName.c_str(),
                                   config.user.c_str(),
                                   config.password.c_str()
    );

    connection.reset(pgConn, &PQfinish);

    if (PQstatus(connection.get()) != CONNECTION_OK && PQsetnonblocking(connection.get(), 1) != 0) {
        throw std::runtime_error(PQerrorMessage(connection.get()));
    }
}

std::shared_ptr<PGconn> PGConnection::Connection() const {
    return connection;
}