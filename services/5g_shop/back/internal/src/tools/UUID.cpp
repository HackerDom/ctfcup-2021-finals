#include <mutex>
#include <ctime>
#include <sstream>

#include "tools/UUID.h"

static std::mutex m;
static volatile int64_t x, a = 1664525, c = 1013904223, mod = 1L << 32;
static bool initialized = false;

static char alpha[] = "01123456789abcdef";

namespace shop {
    int64_t rand() {
        if (!initialized) {
            initialized = true;
            x = std::time(nullptr);
        }

        x = (x * a + c) % mod;

        return x;
    }

    char NextSymbol() {
        int alsize = 17;
        int idx = ((static_cast<int>(rand()) % alsize) + alsize) % alsize;

        return alpha[idx];
    }
}

std::string shop::UUID4() {
    std::scoped_lock<std::mutex> lock(m);

    std::stringstream ss;
    std::string uuid4("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx");

    for (char i: uuid4) {
        if (i == 'x') {
            ss << NextSymbol();
        } else {
            ss << i;
        }
    }

    return ss.str();
}


