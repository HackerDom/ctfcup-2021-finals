#ifndef INTERNAL_DEFER_H
#define INTERNAL_DEFER_H

#include <mutex>
#include <vector>
#include <stack>
#include <functional>
#include <atomic>

namespace shop {
    class Defer {
    public:
        Defer();

        ~Defer();

        Defer(const Defer &other) = delete;

        Defer(Defer &&other) = delete;

        Defer &operator=(const Defer &other) = delete;

        Defer &operator=(Defer &&other) = delete;

        void Add(std::function<void()> &&action);

        template<class F, class ...Args>
        void operator()(F &&action, Args &&...args) {
            auto bind = std::bind(std::forward<F>(action), std::forward<Args>(args)...);

            Add([bind] { bind(); });
        }

        void ExecuteAll();

        [[noreturn]] void *operator new(size_t) {
            throw std::runtime_error("Defer mustn't be allocated on heap");
        }

        [[noreturn]] void operator delete(void*) {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wexceptions"
            throw std::runtime_error("Defer mustn't be allocated on heap");
#pragma clang diagnostic pop
        }

    private:
        std::vector<std::function<void()>> actions;
    };
}

#endif //INTERNAL_DEFER_H
