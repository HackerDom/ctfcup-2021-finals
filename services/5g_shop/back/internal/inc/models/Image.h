#ifndef INTERNAL_IMAGE_H
#define INTERNAL_IMAGE_H

#include <memory>
#include <string>

#include <libpq-fe.h>

namespace shop {
    class Image {
    public:
        Image(int id, int ownerId, std::string filename)
                : id(id), ownerId(ownerId), filename(std::move(filename)) {
        }

        const int id;
        const int ownerId;
        const std::string filename;

        static std::shared_ptr<Image> ReadFromPGResult(PGresult *result, int row = 0) {
            auto idCol = PQfnumber(result, "id");
            auto ownerIdCol = PQfnumber(result, "owner_id");
            auto filenameCol = PQfnumber(result, "filename");

            if (idCol == -1 || ownerIdCol == -1 || filenameCol == -1) {
                throw std::runtime_error("unexpected pg result to read image info");
            }

            return std::make_shared<Image>(
                    std::stoi(std::string(PQgetvalue(result, row, idCol))),
                    std::stoi(std::string(PQgetvalue(result, row, ownerIdCol))),
                    std::string(PQgetvalue(result, row, filenameCol))
            );
        }
    };
}

#endif //INTERNAL_IMAGE_H
