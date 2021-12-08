#ifndef INTERNAL_COMMANDLINEPARSER_H
#define INTERNAL_COMMANDLINEPARSER_H

#include <string>
#include <utility>

namespace shop {
    struct ServerOptions {
    public:
        const std::string address;
        const int port;
    };

    class CommandLineParser {
    public:
        static ServerOptions Parse(int argc, char **argv);
    };
}

#endif //INTERNAL_COMMANDLINEPARSER_H
