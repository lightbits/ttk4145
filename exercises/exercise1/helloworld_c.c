// gcc 4.7.2 +
// gcc -std=gnu99 -Wall -g -o helloworld_c helloworld_c.c -lpthread

#include <pthread.h>
#include <stdio.h>

int i = 0;

// Note the return type: void*
void* thread1Func() {
    for (int k = 0; k < 1000000; k++)
        i++;
    return NULL;
}

void* thread2Func() {
    for (int k = 0; k < 1000000; k++)
        i--;
    return NULL;
}

int main() {
    pthread_t thread1, thread2;
    pthread_create(&thread1, NULL, thread1Func, NULL);
    pthread_create(&thread2, NULL, thread2Func, NULL);
    
    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);

    printf("%d\n", i);

    // On my machine I often get values above 1 000 000 when running this program.
    return 0;
}
