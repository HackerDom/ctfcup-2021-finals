#ifndef INTERNAL_PURCHASE_H
#define INTERNAL_PURCHASE_H

#include <string>
#include <memory>

#include <libpq-fe.h>

namespace shop {
    class Purchase {
    public:
        Purchase(int id, int wareId, int buyerId) : id(id), wareId(wareId), buyerId(buyerId) {
        }

        const int id;
        const int wareId;
        const int buyerId;

        static std::shared_ptr<Purchase> ReadFromPGResult(PGresult *result, int rowNum = 0) {
            auto idCol = PQfnumber(result, "id");
            auto wareIdCol = PQfnumber(result, "ware_id");
            auto buyerIdCol = PQfnumber(result, "buyer_id");

            if (idCol == -1 || wareIdCol == -1 || buyerIdCol == -1) {
                throw std::runtime_error("unexpected pg result to read purchase");
            }

            return std::make_shared<Purchase>(
                    std::stoi(std::string(PQgetvalue(result, rowNum, idCol))),
                    std::stoi(std::string(PQgetvalue(result, rowNum, wareIdCol))),
                    std::stoi(std::string(PQgetvalue(result, rowNum, buyerIdCol)))
            );
        }
    };
}

#endif //INTERNAL_PURCHASE_H
