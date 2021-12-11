#ifndef INTERNAL_SHOPSERVICE_H
#define INTERNAL_SHOPSERVICE_H

#include "services/UsersService.h"
#include "services/ImagesService.h"
#include "services/WaresService.h"
#include "services/PurchasesService.h"

namespace shop {
    class ShopService {
    public:
        ShopService(
                std::shared_ptr<UsersService> usersService,
                std::shared_ptr<ImagesService> imagesService,
                std::shared_ptr<WaresService> waresService,
                std::shared_ptr<PurchasesService> purchasesService
        );

        const std::shared_ptr<UsersService> usersService;
        const std::shared_ptr<ImagesService> imagesService;
        const std::shared_ptr<WaresService> waresService;
        const std::shared_ptr<PurchasesService> purchasesService;
    };
}

#endif //INTERNAL_SHOPSERVICE_H
