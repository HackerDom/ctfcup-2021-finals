#ifndef INTERNAL_MONEYS_H
#define INTERNAL_MONEYS_H

#include <string>

namespace shop {
    int GetServiceFee(int price, const std::string &title, const std::string &description);

    int GetCashback(const std::string &login);
}

#endif //INTERNAL_MONEYS_H
