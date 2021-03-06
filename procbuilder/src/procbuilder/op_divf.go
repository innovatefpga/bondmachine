package procbuilder

import (
	"strconv"
	"strings"
)

// The Divf opcode is both a basic instruction and a template for other instructions.
type Divf struct{}

func (op Divf) Op_get_name() string {
	return "divf"
}

func (op Divf) Op_get_desc() string {
	return "Register divf"
}

func (op Divf) Op_show_assembler(arch *Arch) string {
	opbits := arch.Opcodes_bits()
	result := "divf [" + strconv.Itoa(int(arch.R)) + "(Reg)] [" + strconv.Itoa(int(arch.R)) + "(Reg)]	// Set a register to the division of its value with another register [" + strconv.Itoa(opbits+int(arch.R)+int(arch.R)) + "]\n"
	return result
}

func (op Divf) Op_get_instruction_len(arch *Arch) int {
	opbits := arch.Opcodes_bits()
	return opbits + int(arch.R) + int(arch.R) // The bits for the opcode + bits for a register + bits for another register
}

func (op Divf) Op_instruction_verilog_header(conf *Config, arch *Arch, flavor string) string {
	result := ""
	result += "\treg [31:0] divider_" + arch.Tag + "_input_a;\n"
	result += "\treg [31:0] divider_" + arch.Tag + "_input_b;\n"
	result += "\treg divider_" + arch.Tag + "_input_a_stb;\n"
	result += "\treg divider_" + arch.Tag + "_input_b_stb;\n"
	result += "\treg divider_" + arch.Tag + "_output_z_ack;\n\n"

	result += "\twire [31:0] divider_" + arch.Tag + "_output_z;\n"
	result += "\twire divider_" + arch.Tag + "_output_z_stb;\n"
	result += "\twire divider_" + arch.Tag + "_input_a_ack;\n"
	result += "\twire divider_" + arch.Tag + "_input_b_ack;\n\n"

	result += "\tdivider_" + arch.Tag + " divider_" + arch.Tag + "_inst (divider_" + arch.Tag + "_input_a, divider_" + arch.Tag + "_input_b, divider_" + arch.Tag + "_input_a_stb, divider_" + arch.Tag + "_input_b_stb, divider_" + arch.Tag + "_output_z_ack, clock_signal, reset_signal, divider_" + arch.Tag + "_output_z, divider_" + arch.Tag + "_output_z_stb, divider_" + arch.Tag + "_input_a_ack, divider_" + arch.Tag + "_input_b_ack);\n\n"

	return result
}

func (op Divf) Op_instruction_verilog_state_machine(arch *Arch, flavor string) string {
	rom_word := arch.Max_word()
	opbits := arch.Opcodes_bits()

	reg_num := 1 << arch.R

	result := ""
	result += "					DIVF: begin\n"
	if arch.R == 1 {
		result += "						case (rom_value[" + strconv.Itoa(rom_word-opbits-1) + "])\n"
	} else {
		result += "						case (rom_value[" + strconv.Itoa(rom_word-opbits-1) + ":" + strconv.Itoa(rom_word-opbits-int(arch.R)) + "])\n"
	}
	for i := 0; i < reg_num; i++ {
		result += "						" + strings.ToUpper(Get_register_name(i)) + " : begin\n"

		if arch.R == 1 {
			result += "							case (rom_value[" + strconv.Itoa(rom_word-opbits-int(arch.R)-1) + "])\n"
		} else {
			result += "							case (rom_value[" + strconv.Itoa(rom_word-opbits-int(arch.R)-1) + ":" + strconv.Itoa(rom_word-opbits-int(arch.R)-int(arch.R)) + "])\n"
		}

		for j := 0; j < reg_num; j++ {
			result += "							" + strings.ToUpper(Get_register_name(j)) + " : begin\n"
			result += "							if (!divider_" + arch.Tag + "_output_z_stb) begin\n"
			result += "								divider_" + arch.Tag + "_input_a <= #1 _" + strings.ToLower(Get_register_name(i)) + ";\n"
			result += "								divider_" + arch.Tag + "_input_b <= #1 _" + strings.ToLower(Get_register_name(j)) + ";\n"
			result += "								divider_" + arch.Tag + "_input_a_stb <= #1 1;\n"
			result += "								divider_" + arch.Tag + "_input_b_stb <= #1 1;\n"
			result += "								divider_" + arch.Tag + "_output_z_ack <= #1 0;\n"
			result += "							end\n"
			result += "							else begin\n"
			result += "								_" + strings.ToLower(Get_register_name(i)) + " <= #1 divider_" + arch.Tag + "_output_z;\n"
			result += "								divider_" + arch.Tag + "_output_z_ack <= #1 1;\n"
			result += "								_pc <= #1 _pc + 1'b1 ;\n"
			result += "							end\n"
			result += "								$display(\"DIVF " + strings.ToUpper(Get_register_name(i)) + " " + strings.ToUpper(Get_register_name(j)) + "\");\n"
			result += "							end\n"

		}
		result += "							endcase\n"
		result += "						end\n"
	}
	result += "						endcase\n"
	result += "					end\n"
	return result
}

func (op Divf) Op_instruction_verilog_footer(arch *Arch, flavor string) string {
	// TODO
	return ""
}

func (op Divf) Assembler(arch *Arch, words []string) (string, error) {
	opbits := arch.Opcodes_bits()
	rom_word := arch.Max_word()

	reg_num := 1 << arch.R

	if len(words) != 2 {
		return "", Prerror{"Wrong arguments number"}
	}

	result := ""
	for i := 0; i < reg_num; i++ {
		if words[0] == strings.ToLower(Get_register_name(i)) {
			result += zeros_prefix(int(arch.R), get_binary(i))
			break
		}
	}

	if result == "" {
		return "", Prerror{"Unknown register name " + words[0]}
	}

	partial := ""
	for i := 0; i < reg_num; i++ {
		if words[1] == strings.ToLower(Get_register_name(i)) {
			partial += zeros_prefix(int(arch.R), get_binary(i))
			break
		}
	}

	if partial == "" {
		return "", Prerror{"Unknown register name " + words[1]}
	}

	result += partial

	for i := opbits + 2*int(arch.R); i < rom_word; i++ {
		result += "0"
	}

	return result, nil
}

func (op Divf) Disassembler(arch *Arch, instr string) (string, error) {
	reg_id := get_id(instr[:arch.R])
	result := strings.ToLower(Get_register_name(reg_id)) + " "
	reg_id = get_id(instr[arch.R : 2*int(arch.R)])
	result += strings.ToLower(Get_register_name(reg_id))
	return result, nil
}

// The simulation does nothing
func (op Divf) Simulate(vm *VM, instr string) error {
	// TODO
	vm.Pc = vm.Pc + 1
	return nil
}

// The random genaration does nothing
func (op Divf) Generate(arch *Arch) string {
	// TODO
	return ""
}

func (op Divf) Required_shared() (bool, []string) {
	// TODO
	return false, []string{}
}

func (op Divf) Required_modes() (bool, []string) {
	return false, []string{}
}

func (op Divf) Forbidden_modes() (bool, []string) {
	return false, []string{}
}

func (op Divf) Op_instruction_internal_state(arch *Arch, flavor string) string {
	return ""
}

func (Op Divf) Op_instruction_verilog_reset(arch *Arch, flavor string) string {
	return ""
}

func (Op Divf) Op_instruction_verilog_default_state(arch *Arch, flavor string) string {
	return ""
}

func (Op Divf) Op_instruction_verilog_internal_state(arch *Arch, flavor string) string {
	return ""
}

func (Op Divf) Op_instruction_verilog_extra_modules(arch *Arch, flavor string) ([]string, []string) {
	result := "\n\n"

	result += "//IEEE Floating Point Divider (Single Precision)\n"
	result += "//Copyright (C) Jonathan P Dawson 2013\n"
	result += "//2013-12-12\n"
	result += "//\n"
	result += "module divider_" + arch.Tag + "(\n"
	result += "        input_a,\n"
	result += "        input_b,\n"
	result += "        input_a_stb,\n"
	result += "        input_b_stb,\n"
	result += "        output_z_ack,\n"
	result += "        clk,\n"
	result += "        rst,\n"
	result += "        output_z,\n"
	result += "        output_z_stb,\n"
	result += "        input_a_ack,\n"
	result += "        input_b_ack);\n"
	result += "\n"
	result += "  input     clk;\n"
	result += "  input     rst;\n"
	result += "\n"
	result += "  input     [31:0] input_a;\n"
	result += "  input     input_a_stb;\n"
	result += "  output    input_a_ack;\n"
	result += "\n"
	result += "  input     [31:0] input_b;\n"
	result += "  input     input_b_stb;\n"
	result += "  output    input_b_ack;\n"
	result += "\n"
	result += "  output    [31:0] output_z;\n"
	result += "  output    output_z_stb;\n"
	result += "  input     output_z_ack;\n"
	result += "\n"
	result += "  reg       s_output_z_stb;\n"
	result += "  reg       [31:0] s_output_z;\n"
	result += "  reg       s_input_a_ack;\n"
	result += "  reg       s_input_b_ack;\n"
	result += "\n"
	result += "  reg       [3:0] state;\n"
	result += "  parameter get_a         = 4'd0,\n"
	result += "            get_b         = 4'd1,\n"
	result += "            unpack        = 4'd2,\n"
	result += "            special_cases = 4'd3,\n"
	result += "            normalise_a   = 4'd4,\n"
	result += "            normalise_b   = 4'd5,\n"
	result += "            divide_0      = 4'd6,\n"
	result += "            divide_1      = 4'd7,\n"
	result += "            divide_2      = 4'd8,\n"
	result += "            divide_3      = 4'd9,\n"
	result += "            normalise_1   = 4'd10,\n"
	result += "            normalise_2   = 4'd11,\n"
	result += "            round         = 4'd12,\n"
	result += "            pack          = 4'd13,\n"
	result += "            put_z         = 4'd14;\n"
	result += "\n"
	result += "  reg       [31:0] a, b, z;\n"
	result += "  reg       [23:0] a_m, b_m, z_m;\n"
	result += "  reg       [9:0] a_e, b_e, z_e;\n"
	result += "  reg       a_s, b_s, z_s;\n"
	result += "  reg       guard, round_bit, sticky;\n"
	result += "  reg       [50:0] quotient, divisor, dividend, remainder;\n"
	result += "  reg       [5:0] count;\n"
	result += "\n"
	result += "  always @(posedge clk)\n"
	result += "  begin\n"
	result += "\n"
	result += "    case(state)\n"
	result += "\n"
	result += "      get_a:\n"
	result += "      begin\n"
	result += "        s_input_a_ack <= 1;\n"
	result += "        if (s_input_a_ack && input_a_stb) begin\n"
	result += "          a <= input_a;\n"
	result += "          s_input_a_ack <= 0;\n"
	result += "          state <= get_b;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      get_b:\n"
	result += "      begin\n"
	result += "        s_input_b_ack <= 1;\n"
	result += "        if (s_input_b_ack && input_b_stb) begin\n"
	result += "          b <= input_b;\n"
	result += "          s_input_b_ack <= 0;\n"
	result += "          state <= unpack;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      unpack:\n"
	result += "      begin\n"
	result += "        a_m <= a[22 : 0];\n"
	result += "        b_m <= b[22 : 0];\n"
	result += "        a_e <= a[30 : 23] - 127;\n"
	result += "        b_e <= b[30 : 23] - 127;\n"
	result += "        a_s <= a[31];\n"
	result += "        b_s <= b[31];\n"
	result += "        state <= special_cases;\n"
	result += "      end\n"
	result += "\n"
	result += "      special_cases:\n"
	result += "      begin\n"
	result += "        //if a is NaN or b is NaN return NaN \n"
	result += "        if ((a_e == 128 && a_m != 0) || (b_e == 128 && b_m != 0)) begin\n"
	result += "          z[31] <= 1;\n"
	result += "          z[30:23] <= 255;\n"
	result += "          z[22] <= 1;\n"
	result += "          z[21:0] <= 0;\n"
	result += "          state <= put_z;\n"
	result += "          //if a is inf and b is inf return NaN \n"
	result += "        end else if ((a_e == 128) && (b_e == 128)) begin\n"
	result += "          z[31] <= 1;\n"
	result += "          z[30:23] <= 255;\n"
	result += "          z[22] <= 1;\n"
	result += "          z[21:0] <= 0;\n"
	result += "          state <= put_z;\n"
	result += "        //if a is inf return inf\n"
	result += "        end else if (a_e == 128) begin\n"
	result += "          z[31] <= a_s ^ b_s;\n"
	result += "          z[30:23] <= 255;\n"
	result += "          z[22:0] <= 0;\n"
	result += "          state <= put_z;\n"
	result += "           //if b is zero return NaN\n"
	result += "          if ($signed(b_e == -127) && (b_m == 0)) begin\n"
	result += "            z[31] <= 1;\n"
	result += "            z[30:23] <= 255;\n"
	result += "            z[22] <= 1;\n"
	result += "            z[21:0] <= 0;\n"
	result += "            state <= put_z;\n"
	result += "          end\n"
	result += "        //if b is inf return zero\n"
	result += "        end else if (b_e == 128) begin\n"
	result += "          z[31] <= a_s ^ b_s;\n"
	result += "          z[30:23] <= 0;\n"
	result += "          z[22:0] <= 0;\n"
	result += "          state <= put_z;\n"
	result += "        //if a is zero return zero\n"
	result += "        end else if (($signed(a_e) == -127) && (a_m == 0)) begin\n"
	result += "          z[31] <= a_s ^ b_s;\n"
	result += "          z[30:23] <= 0;\n"
	result += "          z[22:0] <= 0;\n"
	result += "          state <= put_z;\n"
	result += "           //if b is zero return NaN\n"
	result += "          if (($signed(b_e) == -127) && (b_m == 0)) begin\n"
	result += "            z[31] <= 1;\n"
	result += "            z[30:23] <= 255;\n"
	result += "            z[22] <= 1;\n"
	result += "            z[21:0] <= 0;\n"
	result += "            state <= put_z;\n"
	result += "          end\n"
	result += "        //if b is zero return inf\n"
	result += "        end else if (($signed(b_e) == -127) && (b_m == 0)) begin\n"
	result += "          z[31] <= a_s ^ b_s;\n"
	result += "          z[30:23] <= 255;\n"
	result += "          z[22:0] <= 0;\n"
	result += "          state <= put_z;\n"
	result += "        end else begin\n"
	result += "          //Denormalised Number\n"
	result += "          if ($signed(a_e) == -127) begin\n"
	result += "            a_e <= -126;\n"
	result += "          end else begin\n"
	result += "            a_m[23] <= 1;\n"
	result += "          end\n"
	result += "          //Denormalised Number\n"
	result += "          if ($signed(b_e) == -127) begin\n"
	result += "            b_e <= -126;\n"
	result += "          end else begin\n"
	result += "            b_m[23] <= 1;\n"
	result += "          end\n"
	result += "          state <= normalise_a;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      normalise_a:\n"
	result += "      begin\n"
	result += "        if (a_m[23]) begin\n"
	result += "          state <= normalise_b;\n"
	result += "        end else begin\n"
	result += "          a_m <= a_m << 1;\n"
	result += "          a_e <= a_e - 1;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      normalise_b:\n"
	result += "      begin\n"
	result += "        if (b_m[23]) begin\n"
	result += "          state <= divide_0;\n"
	result += "        end else begin\n"
	result += "          b_m <= b_m << 1;\n"
	result += "          b_e <= b_e - 1;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      divide_0:\n"
	result += "      begin\n"
	result += "        z_s <= a_s ^ b_s;\n"
	result += "        z_e <= a_e - b_e;\n"
	result += "        quotient <= 0;\n"
	result += "        remainder <= 0;\n"
	result += "        count <= 0;\n"
	result += "        dividend <= a_m << 27;\n"
	result += "        divisor <= b_m;\n"
	result += "        state <= divide_1;\n"
	result += "      end\n"
	result += "\n"
	result += "      divide_1:\n"
	result += "      begin\n"
	result += "        quotient <= quotient << 1;\n"
	result += "        remainder <= remainder << 1;\n"
	result += "        remainder[0] <= dividend[50];\n"
	result += "        dividend <= dividend << 1;\n"
	result += "        state <= divide_2;\n"
	result += "      end\n"
	result += "\n"
	result += "      divide_2:\n"
	result += "      begin\n"
	result += "        if (remainder >= divisor) begin\n"
	result += "          quotient[0] <= 1;\n"
	result += "          remainder <= remainder - divisor;\n"
	result += "        end\n"
	result += "        if (count == 49) begin\n"
	result += "          state <= divide_3;\n"
	result += "        end else begin\n"
	result += "          count <= count + 1;\n"
	result += "          state <= divide_1;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      divide_3:\n"
	result += "      begin\n"
	result += "        z_m <= quotient[26:3];\n"
	result += "        guard <= quotient[2];\n"
	result += "        round_bit <= quotient[1];\n"
	result += "        sticky <= quotient[0] | (remainder != 0);\n"
	result += "        state <= normalise_1;\n"
	result += "      end\n"
	result += "\n"
	result += "      normalise_1:\n"
	result += "      begin\n"
	result += "        if (z_m[23] == 0 && $signed(z_e) > -126) begin\n"
	result += "          z_e <= z_e - 1;\n"
	result += "          z_m <= z_m << 1;\n"
	result += "          z_m[0] <= guard;\n"
	result += "          guard <= round_bit;\n"
	result += "          round_bit <= 0;\n"
	result += "        end else begin\n"
	result += "          state <= normalise_2;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      normalise_2:\n"
	result += "      begin\n"
	result += "        if ($signed(z_e) < -126) begin\n"
	result += "          z_e <= z_e + 1;\n"
	result += "          z_m <= z_m >> 1;\n"
	result += "          guard <= z_m[0];\n"
	result += "          round_bit <= guard;\n"
	result += "          sticky <= sticky | round_bit;\n"
	result += "        end else begin\n"
	result += "          state <= round;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "      round:\n"
	result += "      begin\n"
	result += "        if (guard && (round_bit | sticky | z_m[0])) begin\n"
	result += "          z_m <= z_m + 1;\n"
	result += "          if (z_m == 24'hffffff) begin\n"
	result += "            z_e <=z_e + 1;\n"
	result += "          end\n"
	result += "        end\n"
	result += "        state <= pack;\n"
	result += "      end\n"
	result += "\n"
	result += "      pack:\n"
	result += "      begin\n"
	result += "        z[22 : 0] <= z_m[22:0];\n"
	result += "        z[30 : 23] <= z_e[7:0] + 127;\n"
	result += "        z[31] <= z_s;\n"
	result += "        if ($signed(z_e) == -126 && z_m[23] == 0) begin\n"
	result += "          z[30 : 23] <= 0;\n"
	result += "        end\n"
	result += "        //if overflow occurs, return inf\n"
	result += "        if ($signed(z_e) > 127) begin\n"
	result += "          z[22 : 0] <= 0;\n"
	result += "          z[30 : 23] <= 255;\n"
	result += "          z[31] <= z_s;\n"
	result += "        end\n"
	result += "        state <= put_z;\n"
	result += "      end\n"
	result += "\n"
	result += "      put_z:\n"
	result += "      begin\n"
	result += "        s_output_z_stb <= 1;\n"
	result += "        s_output_z <= z;\n"
	result += "        if (s_output_z_stb && output_z_ack) begin\n"
	result += "          s_output_z_stb <= 0;\n"
	result += "          state <= get_a;\n"
	result += "        end\n"
	result += "      end\n"
	result += "\n"
	result += "    endcase\n"
	result += "\n"
	result += "    if (rst == 1) begin\n"
	result += "      state <= get_a;\n"
	result += "      s_input_a_ack <= 0;\n"
	result += "      s_input_b_ack <= 0;\n"
	result += "      s_output_z_stb <= 0;\n"
	result += "    end\n"
	result += "\n"
	result += "  end\n"
	result += "  assign input_a_ack = s_input_a_ack;\n"
	result += "  assign input_b_ack = s_input_b_ack;\n"
	result += "  assign output_z_stb = s_output_z_stb;\n"
	result += "  assign output_z = s_output_z;\n"
	result += "\n"
	result += "endmodule\n"

	return []string{"divider"}, []string{result}
}

func (Op Divf) Abstract_Assembler(arch *Arch, words []string) ([]UsageNotify, error) {
	result := make([]UsageNotify, 0)
	return result, nil
}

func (Op Divf) Op_instruction_verilog_extra_block(arch *Arch, flavor string, level uint8, blockname string, objects []string) string {
	result := ""
	switch blockname {
	default:
		result = ""
	}
	return result
}
