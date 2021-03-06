#include <mutex>
#include <ctime>
#include <sstream>

#include "tools/UUID.h"

static std::mutex m;
volatile uint8_t ab[] = {0x41, 0x3c, 0xc6, 0xf3, 0x4e, 0x5f, 0x6d};
volatile uint8_t cb[] = {0x3c, 0x41, 0x6e, 0xc6, 0xf3, 0x6d, 0x5f};
static volatile uint32_t x;
static bool initialized = false;

static char alpha[] = "0123456789abcdef";

namespace shop {
    uint16_t rand() {
        if (!initialized) {
            initialized = true;

            x = std::time(nullptr);
        }

        uint32_t a = (ab[0] << 24) | (ab[2] << 16) | (ab[4] << 8) | ab[6];
        uint32_t c = (cb[0] << 24) | (cb[2] << 16) | (cb[4] << 8) | cb[6];

        x = x * a + c;

        return static_cast<uint16_t>((x & 0xFFFF0000) >> 16);
    }
}

std::string shop::UUID4() {
    std::scoped_lock<std::mutex> lock(m);

    std::stringstream ss;

    int counters[] = {4, 2, 2, 8};

    union {
        uint16_t randvalues[8];
        uint8_t bytes[16];
    };

    randvalues[0] = rand();
    randvalues[1] = rand();
    randvalues[2] = rand();
    randvalues[3] = rand();
    randvalues[4] = rand();
    randvalues[5] = rand();
    randvalues[6] = rand();
    randvalues[7] = rand();

    int p = 0;

    for (auto i = 0; i < 4; ++i) {
        for (int j = 0; j < counters[i]; ++j) {
            uint8_t b = bytes[p++];

            ss << alpha[b & 0x0F] << alpha[(b & 0xF0) >> 4];
        }

        if (i != 3) {
            ss << "-";
        }
    }

    return ss.str();
}
