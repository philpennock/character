// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package transform

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/philpennock/character/internal/table"
)

// This file would be less repetitive, albeit more complex, if it used Reflection to handle the fields
// Revisit the current approach?

// https://en.wikipedia.org/wiki/Mathematical_Alphanumeric_Symbols
type mathVariants struct {
	Normal            rune
	Bold              rune
	Italic            rune
	BoldItalic        rune
	SansNormal        rune
	SansBold          rune
	SansItalic        rune
	SansBoldItalic    rune
	CalligraphyNormal rune
	CalligraphyBold   rune
	FrakturNormal     rune
	FrakturBold       rune
	Monospace         rune
	DoubleStruck      rune
}

var variantList []mathVariants
var variantLookups map[rune]*mathVariants

func init() {
	variantList = []mathVariants{
		{'A', '𝐀', '𝐴', '𝑨', '𝖠', '𝗔', '𝘈', '𝘼', '𝒜', '𝓐', '𝔄', '𝕬', '𝙰', '𝔸'},
		{'B', '𝐁', '𝐵', '𝑩', '𝖡', '𝗕', '𝘉', '𝘽', 'ℬ', '𝓑', '𝔅', '𝕭', '𝙱', '𝔹'},
		{'C', '𝐂', '𝐶', '𝑪', '𝖢', '𝗖', '𝘊', '𝘾', '𝒞', '𝓒', 'ℭ', '𝕮', '𝙲', 'ℂ'},
		{'D', '𝐃', '𝐷', '𝑫', '𝖣', '𝗗', '𝘋', '𝘿', '𝒟', '𝓓', '𝔇', '𝕯', '𝙳', '𝔻'},
		{'E', '𝐄', '𝐸', '𝑬', '𝖤', '𝗘', '𝘌', '𝙀', 'ℰ', '𝓔', '𝔈', '𝕰', '𝙴', '𝔼'},
		{'F', '𝐅', '𝐹', '𝑭', '𝖥', '𝗙', '𝘍', '𝙁', 'ℱ', '𝓕', '𝔉', '𝕱', '𝙵', '𝔽'},
		{'G', '𝐆', '𝐺', '𝑮', '𝖦', '𝗚', '𝘎', '𝙂', '𝒢', '𝓖', '𝔊', '𝕲', '𝙶', '𝔾'},
		{'H', '𝐇', '𝐻', '𝑯', '𝖧', '𝗛', '𝘏', '𝙃', 'ℋ', '𝓗', 'ℌ', '𝕳', '𝙷', 'ℍ'},
		{'I', '𝐈', '𝐼', '𝑰', '𝖨', '𝗜', '𝘐', '𝙄', 'ℐ', '𝓘', 'ℑ', '𝕴', '𝙸', '𝕀'},
		{'J', '𝐉', '𝐽', '𝑱', '𝖩', '𝗝', '𝘑', '𝙅', '𝒥', '𝓙', '𝔍', '𝕵', '𝙹', '𝕁'},
		{'K', '𝐊', '𝐾', '𝑲', '𝖪', '𝗞', '𝘒', '𝙆', '𝒦', '𝓚', '𝔎', '𝕶', '𝙺', '𝕂'},
		{'L', '𝐋', '𝐿', '𝑳', '𝖫', '𝗟', '𝘓', '𝙇', 'ℒ', '𝓛', '𝔏', '𝕷', '𝙻', '𝕃'},
		{'M', '𝐌', '𝑀', '𝑴', '𝖬', '𝗠', '𝘔', '𝙈', 'ℳ', '𝓜', '𝔐', '𝕸', '𝙼', '𝕄'},
		{'N', '𝐍', '𝑁', '𝑵', '𝖭', '𝗡', '𝘕', '𝙉', '𝒩', '𝓝', '𝔑', '𝕹', '𝙽', 'ℕ'},
		{'O', '𝐎', '𝑂', '𝑶', '𝖮', '𝗢', '𝘖', '𝙊', '𝒪', '𝓞', '𝔒', '𝕺', '𝙾', '𝕆'},
		{'P', '𝐏', '𝑃', '𝑷', '𝖯', '𝗣', '𝘗', '𝙋', '𝒫', '𝓟', '𝔓', '𝕻', '𝙿', 'ℙ'},
		{'Q', '𝐐', '𝑄', '𝑸', '𝖰', '𝗤', '𝘘', '𝙌', '𝒬', '𝓠', '𝔔', '𝕼', '𝚀', 'ℚ'},
		{'R', '𝐑', '𝑅', '𝑹', '𝖱', '𝗥', '𝘙', '𝙍', 'ℛ', '𝓡', 'ℜ', '𝕽', '𝚁', 'ℝ'},
		{'S', '𝐒', '𝑆', '𝑺', '𝖲', '𝗦', '𝘚', '𝙎', '𝒮', '𝓢', '𝔖', '𝕾', '𝚂', '𝕊'},
		{'T', '𝐓', '𝑇', '𝑻', '𝖳', '𝗧', '𝘛', '𝙏', '𝒯', '𝓣', '𝔗', '𝕿', '𝚃', '𝕋'},
		{'U', '𝐔', '𝑈', '𝑼', '𝖴', '𝗨', '𝘜', '𝙐', '𝒰', '𝓤', '𝔘', '𝖀', '𝚄', '𝕌'},
		{'V', '𝐕', '𝑉', '𝑽', '𝖵', '𝗩', '𝘝', '𝙑', '𝒱', '𝓥', '𝔙', '𝖁', '𝚅', '𝕍'},
		{'W', '𝐖', '𝑊', '𝑾', '𝖶', '𝗪', '𝘞', '𝙒', '𝒲', '𝓦', '𝔚', '𝖂', '𝚆', '𝕎'},
		{'X', '𝐗', '𝑋', '𝑿', '𝖷', '𝗫', '𝘟', '𝙓', '𝒳', '𝓧', '𝔛', '𝖃', '𝚇', '𝕏'},
		{'Y', '𝐘', '𝑌', '𝒀', '𝖸', '𝗬', '𝘠', '𝙔', '𝒴', '𝓨', '𝔜', '𝖄', '𝚈', '𝕐'},
		{'Z', '𝐙', '𝑍', '𝒁', '𝖹', '𝗭', '𝘡', '𝙕', '𝒵', '𝓩', 'ℨ', '𝖅', '𝚉', 'ℤ'},
		{'a', '𝐚', '𝑎', '𝒂', '𝖺', '𝗮', '𝘢', '𝙖', '𝒶', '𝓪', '𝔞', '𝖆', '𝚊', '𝕒'},
		{'b', '𝐛', '𝑏', '𝒃', '𝖻', '𝗯', '𝘣', '𝙗', '𝒷', '𝓫', '𝔟', '𝖇', '𝚋', '𝕓'},
		{'c', '𝐜', '𝑐', '𝒄', '𝖼', '𝗰', '𝘤', '𝙘', '𝒸', '𝓬', '𝔠', '𝖈', '𝚌', '𝕔'},
		{'d', '𝐝', '𝑑', '𝒅', '𝖽', '𝗱', '𝘥', '𝙙', '𝒹', '𝓭', '𝔡', '𝖉', '𝚍', '𝕕'},
		{'e', '𝐞', '𝑒', '𝒆', '𝖾', '𝗲', '𝘦', '𝙚', 'ℯ', '𝓮', '𝔢', '𝖊', '𝚎', '𝕖'},
		{'f', '𝐟', '𝑓', '𝒇', '𝖿', '𝗳', '𝘧', '𝙛', '𝒻', '𝓯', '𝔣', '𝖋', '𝚏', '𝕗'},
		{'g', '𝐠', '𝑔', '𝒈', '𝗀', '𝗴', '𝘨', '𝙜', 'ℊ', '𝓰', '𝔤', '𝖌', '𝚐', '𝕘'},
		{'h', '𝐡', 'ℎ', '𝒉', '𝗁', '𝗵', '𝘩', '𝙝', '𝒽', '𝓱', '𝔥', '𝖍', '𝚑', '𝕙'},
		{'i', '𝐢', '𝑖', '𝒊', '𝗂', '𝗶', '𝘪', '𝙞', '𝒾', '𝓲', '𝔦', '𝖎', '𝚒', '𝕚'},
		{'j', '𝐣', '𝑗', '𝒋', '𝗃', '𝗷', '𝘫', '𝙟', '𝒿', '𝓳', '𝔧', '𝖏', '𝚓', '𝕛'},
		{'k', '𝐤', '𝑘', '𝒌', '𝗄', '𝗸', '𝘬', '𝙠', '𝓀', '𝓴', '𝔨', '𝖐', '𝚔', '𝕜'},
		{'l', '𝐥', '𝑙', '𝒍', '𝗅', '𝗹', '𝘭', '𝙡', '𝓁', '𝓵', '𝔩', '𝖑', '𝚕', '𝕝'},
		{'m', '𝐦', '𝑚', '𝒎', '𝗆', '𝗺', '𝘮', '𝙢', '𝓂', '𝓶', '𝔪', '𝖒', '𝚖', '𝕞'},
		{'n', '𝐧', '𝑛', '𝒏', '𝗇', '𝗻', '𝘯', '𝙣', '𝓃', '𝓷', '𝔫', '𝖓', '𝚗', '𝕟'},
		{'o', '𝐨', '𝑜', '𝒐', '𝗈', '𝗼', '𝘰', '𝙤', 'ℴ', '𝓸', '𝔬', '𝖔', '𝚘', '𝕠'},
		{'p', '𝐩', '𝑝', '𝒑', '𝗉', '𝗽', '𝘱', '𝙥', '𝓅', '𝓹', '𝔭', '𝖕', '𝚙', '𝕡'},
		{'q', '𝐪', '𝑞', '𝒒', '𝗊', '𝗾', '𝘲', '𝙦', '𝓆', '𝓺', '𝔮', '𝖖', '𝚚', '𝕢'},
		{'r', '𝐫', '𝑟', '𝒓', '𝗋', '𝗿', '𝘳', '𝙧', '𝓇', '𝓻', '𝔯', '𝖗', '𝚛', '𝕣'},
		{'s', '𝐬', '𝑠', '𝒔', '𝗌', '𝘀', '𝘴', '𝙨', '𝓈', '𝓼', '𝔰', '𝖘', '𝚜', '𝕤'},
		{'t', '𝐭', '𝑡', '𝒕', '𝗍', '𝘁', '𝘵', '𝙩', '𝓉', '𝓽', '𝔱', '𝖙', '𝚝', '𝕥'},
		{'u', '𝐮', '𝑢', '𝒖', '𝗎', '𝘂', '𝘶', '𝙪', '𝓊', '𝓾', '𝔲', '𝖚', '𝚞', '𝕦'},
		{'v', '𝐯', '𝑣', '𝒗', '𝗏', '𝘃', '𝘷', '𝙫', '𝓋', '𝓿', '𝔳', '𝖛', '𝚟', '𝕧'},
		{'w', '𝐰', '𝑤', '𝒘', '𝗐', '𝘄', '𝘸', '𝙬', '𝓌', '𝔀', '𝔴', '𝖜', '𝚠', '𝕨'},
		{'x', '𝐱', '𝑥', '𝒙', '𝗑', '𝘅', '𝘹', '𝙭', '𝓍', '𝔁', '𝔵', '𝖝', '𝚡', '𝕩'},
		{'y', '𝐲', '𝑦', '𝒚', '𝗒', '𝘆', '𝘺', '𝙮', '𝓎', '𝔂', '𝔶', '𝖞', '𝚢', '𝕪'},
		{'z', '𝐳', '𝑧', '𝒛', '𝗓', '𝘇', '𝘻', '𝙯', '𝓏', '𝔃', '𝔷', '𝖟', '𝚣', '𝕫'},

		{'0', '𝟎', 0, 0, '𝟢', '𝟬', 0, 0, 0, 0, 0, 0, '𝟶', '𝟘'},
		{'1', '𝟏', 0, 0, '𝟣', '𝟭', 0, 0, 0, 0, 0, 0, '𝟷', '𝟙'},
		{'2', '𝟐', 0, 0, '𝟤', '𝟮', 0, 0, 0, 0, 0, 0, '𝟸', '𝟚'},
		{'3', '𝟑', 0, 0, '𝟥', '𝟯', 0, 0, 0, 0, 0, 0, '𝟹', '𝟛'},
		{'4', '𝟒', 0, 0, '𝟦', '𝟰', 0, 0, 0, 0, 0, 0, '𝟺', '𝟜'},
		{'5', '𝟓', 0, 0, '𝟧', '𝟱', 0, 0, 0, 0, 0, 0, '𝟻', '𝟝'},
		{'6', '𝟔', 0, 0, '𝟨', '𝟲', 0, 0, 0, 0, 0, 0, '𝟼', '𝟞'},
		{'7', '𝟕', 0, 0, '𝟩', '𝟳', 0, 0, 0, 0, 0, 0, '𝟽', '𝟟'},
		{'8', '𝟖', 0, 0, '𝟪', '𝟴', 0, 0, 0, 0, 0, 0, '𝟾', '𝟠'},
		{'9', '𝟗', 0, 0, '𝟫', '𝟵', 0, 0, 0, 0, 0, 0, '𝟿', '𝟡'},
	}

	variantLookups = make(map[rune]*mathVariants, len(variantList)*14)
	for i := range variantList {
		variantLookups[variantList[i].Normal] = &variantList[i]
		variantLookups[variantList[i].Bold] = &variantList[i]
		variantLookups[variantList[i].Italic] = &variantList[i]
		variantLookups[variantList[i].BoldItalic] = &variantList[i]
		variantLookups[variantList[i].SansNormal] = &variantList[i]
		variantLookups[variantList[i].SansBold] = &variantList[i]
		variantLookups[variantList[i].SansItalic] = &variantList[i]
		variantLookups[variantList[i].SansBoldItalic] = &variantList[i]
		variantLookups[variantList[i].CalligraphyNormal] = &variantList[i]
		variantLookups[variantList[i].CalligraphyBold] = &variantList[i]
		variantLookups[variantList[i].FrakturNormal] = &variantList[i]
		variantLookups[variantList[i].FrakturBold] = &variantList[i]
		variantLookups[variantList[i].Monospace] = &variantList[i]
		variantLookups[variantList[i].DoubleStruck] = &variantList[i]
	}
}

func mapRune(r rune, field func(*mathVariants) rune) rune {
	variant, ok := variantLookups[r]
	if !ok {
		return r
	}
	if variant == nil {
		panic("nil variant pointer")
	}
	fr := field(variant)
	if fr == 0 {
		return r
	}
	return fr
}

func toNormal(r rune) rune     { return mapRune(r, func(v *mathVariants) rune { return v.Normal }) }
func toBold(r rune) rune       { return mapRune(r, func(v *mathVariants) rune { return v.Bold }) }
func toItalic(r rune) rune     { return mapRune(r, func(v *mathVariants) rune { return v.Italic }) }
func toBoldItalic(r rune) rune { return mapRune(r, func(v *mathVariants) rune { return v.BoldItalic }) }
func toSansNormal(r rune) rune { return mapRune(r, func(v *mathVariants) rune { return v.SansNormal }) }
func toSansBold(r rune) rune   { return mapRune(r, func(v *mathVariants) rune { return v.SansBold }) }
func toSansItalic(r rune) rune { return mapRune(r, func(v *mathVariants) rune { return v.SansItalic }) }
func toSansBoldItalic(r rune) rune {
	return mapRune(r, func(v *mathVariants) rune { return v.SansBoldItalic })
}
func toCalligraphyNormal(r rune) rune {
	return mapRune(r, func(v *mathVariants) rune { return v.CalligraphyNormal })
}
func toCalligraphyBold(r rune) rune {
	return mapRune(r, func(v *mathVariants) rune { return v.CalligraphyBold })
}
func toFrakturNormal(r rune) rune {
	return mapRune(r, func(v *mathVariants) rune { return v.FrakturNormal })
}
func toFrakturBold(r rune) rune {
	return mapRune(r, func(v *mathVariants) rune { return v.FrakturBold })
}
func toMonospace(r rune) rune { return mapRune(r, func(v *mathVariants) rune { return v.Monospace }) }
func toDoubleStruck(r rune) rune {
	return mapRune(r, func(v *mathVariants) rune { return v.DoubleStruck })
}

var conversions map[string]func(rune) rune

func init() {
	conversions = make(map[string]func(rune) rune, 14)
	conversions["normal"] = toNormal
	conversions["bold"] = toBold
	conversions["italic"] = toItalic
	conversions["bolditalic"] = toBoldItalic
	conversions["sansnormal"] = toSansNormal
	conversions["sansbold"] = toSansBold
	conversions["sansitalic"] = toSansItalic
	conversions["sansbolditalic"] = toSansBoldItalic
	conversions["calligraphynormal"] = toCalligraphyNormal
	conversions["calligraphybold"] = toCalligraphyBold
	conversions["frakturnormal"] = toFrakturNormal
	conversions["frakturbold"] = toFrakturBold
	conversions["monospace"] = toMonospace
	conversions["doublestruck"] = toDoubleStruck
}

func mathListCommands(w io.Writer, verbose bool, args []string) error {
	avail := make([]string, 0, len(conversions))
	for k := range conversions {
		avail = append(avail, k)
	}
	sort.Strings(avail)
	if !verbose || !table.Supported() {
		for _, item := range avail {
			fmt.Fprintf(w, "  %q\n", item)
		}
		return nil
	}
	t := table.New()
	columns := make([]any, 0, 3)
	columns = append(columns, "Name", "Rendered Name")
	var exemplar string
	if len(args) > 0 {
		columns = append(columns, "Exemplar")
		exemplar = strings.Join(args, " ")
	}
	t.AddHeaders(columns...)
	for _, item := range avail {
		row := make([]any, 2, 3)
		row[0] = item
		row[1] = strings.Map(conversions[item], item)
		if exemplar != "" {
			row = append(row, strings.Map(conversions[item], exemplar))
		}
		t.AddRow(row...)
	}
	_, err := fmt.Fprint(w, t.Render())
	return err
}

// TransformMath converts each argument to a mathematical letter variant and
// returns the results joined by spaces.  target names a conversion type (e.g.
// "bold", "italic", "frakturnormal"); an empty string is treated as "normal".
// It is the exported API for non-Cobra callers (e.g. the MCP server).
func TransformMath(args []string, target string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	want := strings.Map(func(r rune) rune {
		if unicode.IsLower(r) {
			return r
		}
		return -1
	}, strings.ToLower(target))
	if len(want) == 0 {
		want = "normal"
	}
	convert, ok := conversions[want]
	if !ok {
		return "", fmt.Errorf("math: unknown conversion %q", target)
	}
	output := make([]string, len(args))
	for argI := range args {
		output[argI] = strings.Map(convert, args[argI])
	}
	return strings.Join(output, " "), nil
}

var mathSubcommand = transformCobraCommand{
	Use:   "math",
	Short: "map characters between math variants",
	Transformer: func(args []string) (string, error) {
		if len(args) == 0 {
			return "", nil
		}
		want := strings.Map(func(r rune) rune {
			if unicode.IsLower(r) {
				return r
			}
			return -1
		}, strings.ToLower(flags.target))
		if len(want) == 0 {
			want = "normal"
		}
		convert, ok := conversions[want]
		if !ok {
			return "", fmt.Errorf("math: unknown conversion %q\n", flags.target)
		}

		output := make([]string, len(args))
		for argI := range args {
			output[argI] = strings.Map(convert, args[argI])
		}
		return strings.Join(output, " "), nil
	},
	List: mathListCommands,
}
