#include <mutex>
#include <ctime>
#include <sstream>

#include "tools/UUID.h"

static std::mutex m;
static volatile uint32_t x, a = 1103515245, c = 1013904223;
static bool initialized = false;

static char alpha[] = "0123456789abcdef";

namespace shop {
    uint16_t rand() {
        if (!initialized) {
            initialized = true;

            x = std::time(nullptr);
        }

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
