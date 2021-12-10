#ifndef INTERNAL_STRINGS_H
#define INTERNAL_STRINGS_H

#include <sstream>
#include <string>
#include <vector>
#include <cstdlib>
#include <memory>
#include <stdexcept>

namespace shop {
    std::string GetEnv(const char *name);

    int GetIntEnv(const char *name);

    std::vector<std::string> Split(const std::string &s, const std::string &delimiter);

    std::string &Ltrim(std::string &s);

    std::string &Rtrim(std::string &s);

    std::string &Trim(std::string &s);

    template<typename ... Args>
    std::string Format(const char *format, Args ... args) {
        int sizeS = std::snprintf(nullptr, 0, format, args ...) + 1;
        if (sizeS <= 0) {
            throw std::runtime_error("Error during formatting.");
        }

        auto size = static_cast<size_t>(sizeS);
        auto buf = std::make_unique<char[]>(size);

        std::snprintf(buf.get(), size, format, args ...);

        return {buf.get(), buf.get() + size - 1};
    }

    template<size_t N>
    struct XorString {
    private:
        const char _key;
        std::array<char, N + 1> _encrypted;

        [[nodiscard]] constexpr char enc(char c) const {
            return c ^ _key;
        }

        [[nodiscard]] char dec(char c) const {
            return c ^ _key;
        }

    public:
        template<size_t... Is>
        constexpr XorString(const char *str, std::index_sequence<Is...>) : _key(0x42), _encrypted{enc(str[Is])...} {
        }

        auto decrypt() {
            for (size_t i = 0; i < N; ++i) {
                _encrypted[i] = dec(_encrypted[i]);
            }
            _encrypted[N] = '\0';
            return _encrypted.data();
        }

#define HiddenStr(s) []{ constexpr XorString<sizeof(s)/sizeof(char) - 1> expr( s, std::make_index_sequence< sizeof(s)/sizeof(char) - 1>() ); return expr; }().decrypt()
    };
}

#endif //INTERNAL_STRINGS_H
