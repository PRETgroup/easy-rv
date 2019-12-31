.PHONY: default c_mon verilog_mon $(PROJECT) c_build
.PRECIOUS: %.xml 

# run this makefile with the following options
# make [c_mon] [c_build] [run_(c/e)bmc] PROJECT=XXXXX FILE=YYYYY
#   PROJECT = name of project directory
#   FILE    = name of file within project directory (default = PROJECT, e.g. example/ab5/ab5.whatever)
#
#   c_mon: make a C monitor for the project
#   c_build: compile the C monitor with a main file (this will need to be provided manually)
#   run_cbmc: check the compiled C monitor to ensure correctness
#
# make [verilog_mon] [run_ebmc] PROJECT=XXXXX
#   verilog_mon: make a Verilog monitor for the project
#   run_ebmc: check the compiled Verilog monitor to ensure correctness

FILE ?= $(PROJECT)
PARSEARGS ?=

default: easy-rv-c easy-rv-parser

#convert C build instruction to C target
c_mon: default $(PROJECT)

#convert verilog build instruction to verilog target
verilog_mon: $(PROJECT)_V

easy-rv-c: rvc/* rvdef/*
	go get github.com/PRETgroup/stcompilerlib
	go build -o easy-rv-c -i ./rvc/main

easy-rv-parser: rvparser/* rvdef/*
	go get github.com/PRETgroup/stcompilerlib
	go build -o easy-rv-parser -i ./rvparser/main

run_cbmc: default 
	cbmc example/$(PROJECT)/cbmc_main_$(PROJECT).c example/$(PROJECT)/F_$(PROJECT).c

run_ebmc: default 
	#$(foreach file,$(wildcard example/$(PROJECT)/*.sv), time --format="took %E" ebmc $(file) --k-induction --trace --top F_combinatorialVerilog_$(word 3,$(subst _, ,$(basename $(notdir $(file)))));)
	time --format="took %E" ebmc example/$(PROJECT)/test_F_$(FILE).sv --k-induction --trace --module F_combinatorialVerilog_$(FILE)
	#ebmc $^ --k-induction --trace

#convert $(PROJECT) into the C binary name
$(PROJECT): ./example/$(PROJECT)/$(FILE).c

#generate the C sources from the erv files
%.c: %.xml
	./easy-rv-c -i $^ -o example/$(PROJECT)

#convert $(PROJECT)_V into the verilog names
$(PROJECT)_V: ./example/$(PROJECT)/$(FILE).sv

#generate the xml from the erv files
%.xml: %.erv
	./easy-rv-parser $(PARSEARGS) -i $^ -o $@

#generate the Verilog sources from the xml files
%.sv: %.xml
	./easy-rv-c -i $^ -o example/$(PROJECT) -l=verilog

#Bonus: C compilation: convert $(PROJECT) into the C binary name
c_build: example_$(PROJECT)

#generate the C binary from the C sources
example_$(PROJECT): example/$(PROJECT)/$(PROJECT)_main.c example/$(PROJECT)/F_$(PROJECT).c
	gcc example/$(PROJECT)/$(PROJECT)_main.c example/$(PROJECT)/F_$(PROJECT).c -o example_$(PROJECT)

#Bonus: C assembly
c_asm: example/$(PROJECT)/F_$(PROJECT).c
	gcc -S example/$(PROJECT)/F_$(PROJECT).c -o example/$(PROJECT)/F_$(PROJECT).s

clean: clean_examples
	rm -f easy-rv-c
	rm -f easy-rv-parser
	go get -u github.com/PRETgroup/stcompilerlib

clean_examples:
	rm -f example_*
	rm -f ./example/*/F_*
	rm -f ./example/*/*.h
	rm -f ./example/*/*.v
	rm -f ./example/*/*.sv
	rm -f ./example/*/*.xml