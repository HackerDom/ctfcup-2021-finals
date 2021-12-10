#ifndef INTERNAL_IMAGESSERVICE_H
#define INTERNAL_IMAGESSERVICE_H

#include "models/Image.h"
#include "tools/Result.h"
#include "models/PGConnectionPool.h"

namespace shop {
    class ImagesService {
    public:
        explicit ImagesService(std::shared_ptr<PGConnectionPool> pgConnectionPool);

        Result<std::shared_ptr<Image>> Get(int id);

        Result<std::shared_ptr<Image>> FindBySha256(const std::string &sha256);

        Result<std::shared_ptr<Image>> Create(int ownerId, const std::string &filename, const std::string &sha256);

        Result<std::vector<std::shared_ptr<Image>>> ListOfUser(int ownerId);

    private:
        std::shared_ptr<PGConnectionPool> pgConnectionPool;
    };

}

#endif //INTERNAL_IMAGESSERVICE_H
