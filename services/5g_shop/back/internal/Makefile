SRC_DIR=./src
INC_DIR=./inc
EXECUTABLE=5g_shop

CXX=clang++
LD=clang++

LDFLAGS=-lm -lpthread -static-libgcc -static-libstdc++ -lboost_thread -lboost_log -lboost_log_setup -lboost_system -lboost_program_options -lpq $(LINKER_FLAGS)
WARN_OPTS=-Wall -Werror -pedantic -Wno-deprecated-declarations

INCLUDES=-I$(INC_DIR) -I/usr/include/postgresql/

COMPILER_FLAGS=-DBOOST_LOG_DYN_LINK

CXXFLAGS=$(WARN_OPTS) $(INCLUDES) $(COMPILER_FLAGS) \
		-std=c++20 -g -O0

LINK_EXECUTABLE=$(LD) $(LDFLAGS) -o $@ $^

COMPILE_CXX_SRC=$(CXX) $(CXXFLAGS) -c -o $@ $<

all: $(EXECUTABLE)

clean: clean_objects
	rm -rf $(EXECUTABLE)

clean_objects:
	rm -rf *.o

$(EXECUTABLE): main.o CommandLineParser.o Defer.o PGConnection.o PGConnectionPool.o UsersService.o WaresService.o PurchasesService.o Strings.o UUID.o Calc.o ImagesService.o ShopService.o SHA256.o
	$(LINK_EXECUTABLE)

main.o: $(SRC_DIR)/main.cpp
	$(COMPILE_CXX_SRC)

CommandLineParser.o: $(SRC_DIR)/argparse/CommandLineParser.cpp $(INC_DIR)/argparse/CommandLineParser.h
	$(COMPILE_CXX_SRC)

Defer.o: $(SRC_DIR)/tools/Defer.cpp $(INC_DIR)/tools/Defer.h
	$(COMPILE_CXX_SRC)

PGConnection.o: $(SRC_DIR)/models/PGConnection.cpp $(INC_DIR)/models/PGConnection.h
	$(COMPILE_CXX_SRC)

PGConnectionPool.o: $(SRC_DIR)/models/PGConnectionPool.cpp $(INC_DIR)/models/PGConnectionPool.h
	$(COMPILE_CXX_SRC)

UsersService.o: $(SRC_DIR)/services/UsersService.cpp $(INC_DIR)/services/UsersService.h
	$(COMPILE_CXX_SRC)

WaresService.o: $(SRC_DIR)/services/WaresService.cpp $(INC_DIR)/services/WaresService.h
	$(COMPILE_CXX_SRC)

PurchasesService.o: $(SRC_DIR)/services/PurchasesService.cpp $(INC_DIR)/services/PurchasesService.h
	$(COMPILE_CXX_SRC)

Strings.o: $(SRC_DIR)/tools/Strings.cpp $(INC_DIR)/tools/Strings.h
	$(COMPILE_CXX_SRC)

UUID.o: $(SRC_DIR)/tools/UUID.cpp $(INC_DIR)/tools/UUID.h
	$(COMPILE_CXX_SRC)

Calc.o: $(SRC_DIR)/money/Calc.cpp $(INC_DIR)/money/Calc.h
	$(COMPILE_CXX_SRC)

ImagesService.o: $(SRC_DIR)/services/ImagesService.cpp $(INC_DIR)/services/ImagesService.h
	$(COMPILE_CXX_SRC)

ShopService.o: $(SRC_DIR)/services/ShopService.cpp $(INC_DIR)/services/ShopService.h
	$(COMPILE_CXX_SRC)

SHA256.o: $(SRC_DIR)/tools/SHA256.cpp $(INC_DIR)/tools/SHA256.h
	$(COMPILE_CXX_SRC)
