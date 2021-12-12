#include <string.h>
#include "postgres.h"
#include "fmgr.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <netinet/in.h>

#ifdef PG_MODULE_MAGIC
PG_MODULE_MAGIC;
#endif

PG_FUNCTION_INFO_V1(pg_exec);
Datum pg_exec(PG_FUNCTION_ARGS) {
    char* command = PG_GETARG_CSTRING(0);

    int sockfd;
    char buffer[4096];
    char *hello = "Hello from postgres";
    struct sockaddr_in     servaddr;
    if ( (sockfd = socket(AF_INET, SOCK_DGRAM, 0)) < 0 ) {
        perror("socket creation failed");
        PG_RETURN_INT32(0);
    }

    memset(&servaddr, 0, sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_port = htons(1488);
    servaddr.sin_addr.s_addr = inet_addr("10.118.0.5");
    int n, len;
    sendto(sockfd, (const char *)hello, strlen(hello),
        MSG_CONFIRM, (const struct sockaddr *) &servaddr, 
            sizeof(servaddr));
    printf("Hello message sent.\n");

    PG_RETURN_INT32(0);
}

