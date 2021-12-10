#include <algorithm>
#include <functional>
#include <cctype>
#include <locale>

#include "tools/Strings.h"

using namespace shop;

std::string shop::GetEnv(const char *name) {
    auto *value = std::getenv(name);

    if (value == nullptr) {
        throw std::runtime_error(Format("environment variable '%s' required", name));
    }

    return {value};
}

int shop::GetIntEnv(const char *name) {
    return std::stoi(GetEnv(name));
}

std::vector<std::string> shop::Split(const std::string &s, const std::string &delimiter) {
    size_t startIdx = 0, endIdx, delimSize = delimiter.length();
    std::string token;
    std::vector<std::string> res;

    while ((endIdx = s.find(delimiter, startIdx)) != std::string::npos) {
        token = s.substr(startIdx, endIdx - startIdx);
        startIdx = endIdx + delimSize;
        res.push_back(token);
    }

    res.push_back(s.substr(startIdx));

    return res;
}

std::string &shop::Ltrim(std::string &s) {
    s.erase(s.begin(), std::find_if(s.begin(), s.end(),
                                    std::not1(std::ptr_fun<int, int>(std::isspace))));
    return s;
}

std::string &shop::Rtrim(std::string &s) {
    s.erase(std::find_if(s.rbegin(), s.rend(),
                         std::not1(std::ptr_fun<int, int>(std::isspace))).base(), s.end());
    return s;
}

std::string &shop::Trim(std::string &s) {
    return Ltrim(Rtrim(s));
}