#ifndef INTERNAL_RESULT_H
#define INTERNAL_RESULT_H

#include <string>
#include <utility>

namespace shop {
    template<class T>
    class Result {
    public:
        const bool success;
        const T value;
        const std::string message;

        static Result ofSuccess(T value) {
            return {true, std::move(value), ""};
        }

        static Result ofSuccess(T value, std::string message) {
            return {true, std::move(value), std::move(message)};
        }

        static Result ofError() {
            return {false, T(), ""};
        }

        static Result ofError(std::string message) {
            return {false, T(), std::move(message)};
        }

    private:
        Result(bool success, T value, std::string message)
                : success(success), value(std::move(value)), message(std::move(message)) {
        }
    };

    typedef Result<void> JustResult;

    template<>
    class Result<void> {
    public:
        const bool success;
        const std::string message;

        static Result ofSuccess() {
            return {true, ""};
        }

        static Result ofSuccess(std::string message) {
            return {true, std::move(message)};
        }

        static Result ofError() {
            return {false, ""};
        }

        static Result ofError(std::string message) {
            return {false, std::move(message)};
        }

    private:
        Result(bool success, std::string message) : success(success), message(std::move(message)) {
        }
    };
}

#endif //INTERNAL_RESULT_H
