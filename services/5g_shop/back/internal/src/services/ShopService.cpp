#include "services/ShopService.h"

using namespace shop;

ShopService::ShopService(
        std::shared_ptr<UsersService> usersService,
        std::shared_ptr<ImagesService> imagesService,
        std::shared_ptr<WaresService> waresService,
        std::shared_ptr<PurchasesService> purchasesService)
        : usersService(std::move(usersService)),
          imagesService(std::move(imagesService)),
          waresService(std::move(waresService)),
          purchasesService(std::move(purchasesService)) {
}