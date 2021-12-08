#include <string>
#include <vector>
#include <array>

extern "C" {
    int Compute(const char *textt);
    int GetCashback(const char *login);
    int GetServiceFee(int price, const char *s1, const char *s2);
}

int Compute(const char *textt) {
    std::string text(textt);
    std::vector<uint8_t> textCopy(text.size() + 100);

    for (char i: text) {
        textCopy.push_back(i);
    }

    while (textCopy.size() % 64 != 0) {
        textCopy.push_back('@');
    }

    uint32_t totalState[8] = {
            0x7f0b2741, 0x14152a41, 0x934f3441, 0xdd203741, 0x73463b41, 0x21a33c41, 0x0d4e4341, 0x7e9c4441
    };

    constexpr std::array<uint32_t, 64> K = {
            0x15dbc940, 0x9973cb40, 0xce05cd40, 0xf091ce40,
            0xac14cf40, 0x0f99d040, 0xe098d140, 0xe017d240,
            0x0e8ad440, 0x34dfd740, 0x2fced840, 0xe844d940,
            0xd730da40, 0x505bdd40, 0xaaafde40, 0x0ddee040,
            0x704ce140, 0xf327e240, 0x216ee340, 0x6f1be540,
            0x5559e640, 0xd893e740, 0xad63e840, 0xbb98e940,
            0xe52feb40, 0x6ef9eb40, 0x8388ed40, 0x1d74ef40,
            0x7ed5ef40, 0xd2b7f140, 0x6317f240, 0x5634f340,
            0xdcf0f340, 0x8909f540, 0xe77bf640, 0x7733f740,
            0xd98ef740, 0xd544f840, 0x9e60fa40, 0x7cc3fb40,
            0x7773fc40, 0x92d0fd40, 0xba7dfe40, 0xc07fff40,
            0xe3be0041, 0x02e90041, 0x575f0241, 0x45da0241,
            0x2ca50341, 0xc01d0441, 0x79950441, 0x32bd0441,
            0xce330541, 0xacf70541, 0x206c0641, 0xccdf0641,
            0x2f060741, 0xd5780741, 0xbaea0741, 0x3f360841,
            0xe25b0841, 0x043c0941, 0xa2f40941, 0x53190a41
    };

    static constexpr std::array<uint32_t, 32> M = {
            0x7c620a41, 0xa8cf0a41, 0x2a3c0b41, 0x2f600b41,
            0xcf360c41, 0xc87d0c41, 0xb8e70c41, 0x03740d41,
            0xdc210e41, 0xc0ab0e41, 0xaa560f41, 0x42de0f41,
            0x4d431041, 0xcba71041, 0x7dea1041, 0x2a6f1141,
            0x0dd21141, 0xb0131241, 0x45961241, 0x39d71241,
            0xcfb81341, 0x4e581441, 0xf4151541, 0x61351541,
            0xbcd11541, 0xdbf01541, 0xf12e1641, 0xe94d1641,
            0x02e81641, 0xb3bd1741, 0x53fa1741, 0x91181841
    };

    uint8_t data[64];
    uint32_t blockLen = 0;
    auto *str = new uint8_t[32];

    for (uint8_t ii: textCopy) {
        data[blockLen++] = ii;

        if (blockLen == 64) {
            uint32_t maj, xorA, ch, xorE, sum, newA, newE, m[64];
            uint32_t state[8];

            for (uint8_t i = 0, j = 0; i < 16; i++, j += 4) {
                m[i] = (data[j] << 24) | (data[j + 1] << 16) | (data[j + 2] << 8) | (data[j + 3]);
            }

            for (uint8_t k = 16; k < 64; k++) {
                m[k] = (((m[k - 2] >> 17) | (m[k - 2] << (32 - 17))) ^ ((m[k - 2] >> 19) | (m[k - 2] << (32 - 19))) ^
                        (m[k - 2] >> 10)) + m[k - 7] +
                       (((m[k - 15] >> 7) | (m[k - 15] << (32 - 7))) ^ ((m[k - 15] >> 18) | (m[k - 15] << (32 - 18))) ^
                        (m[k - 15] >> 3)) + m[k - 16];
            }

            for (uint8_t i = 0; i < 8; i++) {
                state[i] = totalState[i];
            }

            for (uint8_t i = 0; i < 64; i++) {
                maj = (state[0] & (state[1] | state[2])) | (state[1] & state[2]);
                xorA = ((state[0] >> 2) | (state[0] << (32 - 2))) ^ ((state[0] >> 13) | (state[0] << (32 - 13))) ^
                       ((state[0] >> 22) | (state[0] << (32 - 22)));

                ch = (state[4] & state[5]) ^ (~state[4] & state[6]);

                xorE = ((state[4] >> 6) | (state[4] << (32 - 6))) ^ ((state[4] >> 11) | (state[4] << (32 - 11))) ^
                       ((state[4] >> 25) | (state[4] << (32 - 25)));

                sum = m[i] + K[i] + state[7] + ch + xorE;
                newA = xorA + maj + sum;
                newE = state[3] + sum;

                state[7] = state[6];
                state[6] = state[5];
                state[5] = state[4];
                state[4] = newE;
                state[3] = state[2];
                state[2] = state[1];
                state[1] = state[0];
                state[0] = newA;
            }

            for (uint8_t i = 0; i < 8; i++) {
                totalState[i] += state[i];
            }

            blockLen = 0;
        }
    }

    for (uint8_t i = 0; i < 4; i++) {
        for (uint8_t j = 0; j < 8; j++) {
            str[i + (j * 4)] = (totalState[j] >> (24 - i * 8)) & 0xff;
        }
    }

    int result = 0;

    for (uint8_t i = 0; i < 32; i++) {
        result += static_cast<int>(M[i] * static_cast<uint32_t>(str[i]));
    }

    delete[] str;

    return std::abs(result);
}

int GetCashback(const char *login) {
    return Compute(login) % 30;
}

int GetServiceFee(int price, const char *title, const char *description) {
    return Compute(std::string(std::to_string(price) + title + description).c_str()) % (price / 5 + 1);
}

