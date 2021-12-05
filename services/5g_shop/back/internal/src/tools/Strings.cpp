#include "tools/Strings.h"

using namespace shop;

std::string shop::GetEnv(const char *name) {
    auto *value = std::getenv(name);

    if (value == nullptr) {
        throw std::runtime_error(format("environment variable '%s' required", name));
    }

    return {value};
}

int shop::GetIntEnv(const char *name) {
    return std::stoi(GetEnv(name));
}
