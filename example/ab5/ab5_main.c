#include "F_ab5.h"
#include <stdio.h>
#include <stdint.h>

void print_data(uint32_t count, uint8_t status, io_ab5_t io) {
    printf("Tick %7d: A:%d, B:%d, STATUS:%d\r\n", count, io.A, io.B, status);
}

int main() {
    monitorvars_ab5_t enf;
    io_ab5_t io;
    
    ab5_init_all_vars(&enf, &io);

    uint32_t count = 0;
    while(count++ < 40) {
        io.A = false;
        if(count == 3) io.A = true;
        if(count == 7) io.A = true;
        if(count == 11) io.A = true;
        if(count == 12) io.A = true;
        
        ab5_run_via_monitor(&enf, &io);

        print_data(count, ab5_check_rv_status_AB5(&enf), io);
    }
}

void ab5_run(io_ab5_t *io) {
    static bool pre_A = false;
    //do nothing
    io->B = pre_A;
    pre_A = io->A;
}

