#include <cstdlib>
#include <iostream>

#include "tools/Defer.h"

using namespace shop;

static thread_local std::stack<Defer *> *thisThreadDefers;

static std::mutex m;
static bool atExitInitialized = false;
static std::vector<std::stack<Defer *> *> allThreadsDefersLists;

void DoDeferAllOnExit() {
    std::scoped_lock<std::mutex> lock(m);

    for (auto &threadDefer: allThreadsDefersLists) {
        while (!threadDefer->empty()) {
            threadDefer->top()->ExecuteAll();
            threadDefer->pop();
        }
    }
}

std::stack<Defer *> &GetThreadDefers() {
    if (thisThreadDefers == nullptr) {
        thisThreadDefers = new std::stack<Defer *>();

        {
            std::scoped_lock<std::mutex> lock(m);

            allThreadsDefersLists.push_back(thisThreadDefers);

            if (!atExitInitialized) {
                atExitInitialized = true;
                std::atexit(DoDeferAllOnExit);
            }
        }
    }

    return *thisThreadDefers;
}

Defer::Defer() {
    GetThreadDefers().push(this);
}

Defer::~Defer() {
    ExecuteAll();

    GetThreadDefers().pop();
}

void Defer::ExecuteAll() {
    for (auto &a: actions) {
        a();
    }

    actions.clear();
}

void Defer::Add(std::function<void()> &&action) {
    actions.push_back(std::forward<std::function<void()>>(action));
}
