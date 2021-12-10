#ifndef INTERNAL_WARE_H
#define INTERNAL_WARE_H

#include <string>

#include <libpq-fe.h>

namespace shop {
    class Ware {
    public:
        Ware(int id, int sellerId, std::string title, std::string description, int price, int serviceFee, int imageId)
                : id(id),
                  sellerId(sellerId),
                  title(std::move(title)),
                  description(std::move(description)),
                  price(price),
                  serviceFee(serviceFee),
                  imageId(imageId) {
        }

        const int id;
        const int sellerId;
        const std::string title;
        const std::string description;
        const int price;
        const int serviceFee;
        const int imageId;

        static std::shared_ptr<Ware> ReadFromPGResult(PGresult *result, int rowNum = 0) {
            auto idCol = PQfnumber(result, "id");
            auto sellerIdCol = PQfnumber(result, "seller_id");
            auto titleCol = PQfnumber(result, "title");
            auto descriptionCol = PQfnumber(result, "description");
            auto priceCol = PQfnumber(result, "price");
            auto serviceFeeCol = PQfnumber(result, "service_fee");
            auto imageIdCol = PQfnumber(result, "image_id");

            if (sellerIdCol == -1 || titleCol == -1 || descriptionCol == -1 || priceCol == -1 || idCol == -1
                || serviceFeeCol == -1 || imageIdCol == -1) {
                throw std::runtime_error("unexpected pg result to read ware");
            }

            return std::make_shared<Ware>(
                    std::stoi(std::string(PQgetvalue(result, rowNum, idCol))),
                    std::stoi(std::string(PQgetvalue(result, rowNum, sellerIdCol))),
                    std::string(PQgetvalue(result, rowNum, titleCol)),
                    std::string(PQgetvalue(result, rowNum, descriptionCol)),
                    std::stoi(std::string(PQgetvalue(result, rowNum, priceCol))),
                    std::stoi(std::string(PQgetvalue(result, rowNum, serviceFeeCol))),
                    std::stoi(std::string(PQgetvalue(result, rowNum, imageIdCol)))
            );
        }

    };
}

#endif //INTERNAL_WARE_H
