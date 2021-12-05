#include <iostream>
#include <cstdlib>

#include "argparse/CommandLineParser.h"

#include <boost/program_options.hpp>

using namespace shop;

namespace po = boost::program_options;

ServerOptions Parse(int argc, char **argv) {
    po::options_description generalOptions("General options");
    std::string type, address;
    int port;
    generalOptions.add_options()
            ("help", "Show help")
            ("address", po::value<std::string>(&address)->required(), "Address to bind for listening")
            ("port", po::value<int>(&port)->required(), "Port to bind for listening");

    po::variables_map variablesMap;
    auto parsed = po::command_line_parser(argc, argv).options(generalOptions).allow_unregistered().run();
    po::store(parsed, variablesMap);

    if (variablesMap.count("help")) {
        std::cout << generalOptions << std::endl;

        std::exit(EXIT_SUCCESS);
    }

    po::notify(variablesMap);

    return {
            address,
            port
    };
}

ServerOptions CommandLineParser::Parse(int argc, char **argv) {
    try {
        return ::Parse(argc, argv);
    } catch (std::exception &e) {
        std::cout << e.what() << std::endl
                  << "See --help for help" << std::endl;
        std::exit(EXIT_FAILURE);
    } catch (...) {
        std::cout << "Unknown error" << std::endl
                  << "See --help for help" << std::endl;
        std::exit(EXIT_FAILURE);
    }
}
