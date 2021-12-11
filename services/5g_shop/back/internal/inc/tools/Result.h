#ifndef INTERNAL_RESULT_H
#define INTERNAL_RESULT_H

#include <string>
#include <utility>
#include <functional>

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

        template <class TNext>
        Result<TNext> then(std::function<Result<TNext>(const T&)> map) {
            if (!success) {
                return Result<TNext>::ofError(message);
            }

            return map(value);
        }

    private:
        Result(bool success, T value, std::string message)
                : success(success), value(std::move(value)), message(std::move(message)) {
        }
    };
}

#endif //INTERNAL_RESULT_H
