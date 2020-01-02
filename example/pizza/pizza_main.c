#include "F_pizza.h"
#include <stdio.h>
#include <stdint.h>

void print_data(uint32_t count, monitorvars_pizza_t mon, io_pizza_t io) {
    printf("Tick %6d: temp:%5d C, STATE: (%4d, %4ld, %4ld), STATUS (Can Eat?):", count, io.t, mon._policy_FoodSafety_state, mon.xloc, mon.xage);
    switch(pizza_check_rv_status_FoodSafety(&mon)) {
        case 0: printf("TRUE"); break;
        case 1: printf("CURRENTLY TRUE"); break;
        case 2: printf("CURRENTLY FALSE"); break;
        default: printf("FALSE"); break;
    }
    printf("\r\n");
}

int main() {
    monitorvars_pizza_t mon;
    io_pizza_t io;
    
    pizza_init_all_vars(&mon, &io);

    uint32_t count = 0;
    io.t = 0;
    while(count++ < 100) {
        if(count == 2) io.t = 5; //start heating the pizza
        if(count == 5) io.t = 10;
        if(count == 15) io.t = 30;
        if(count == 20) io.t = 60;
        if(count == 25) io.t = 180; //pizza fully heated, remove from oven
        if(count == 30) io.t = 140;
        if(count == 35) io.t = 40; //pizza is cooled enough to eat

        //wait here until pizza expires

        pizza_run_via_monitor(&mon, &io);

        print_data(count, mon, io);
    }
}

void pizza_run(io_pizza_t *io) {
    //nothing to do, the pizza has no operations
}

