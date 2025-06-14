package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

type Variant struct {
	Glyph    string `json:"glyph"`
	Punycode string `json:"punycode"`
}

type LetterOutput struct {
	Letter   string    `json:"letter"`
	Variants []Variant `json:"variants"`
}

type WordOutput struct {
	Word     string   `json:"word"`
	Variants []string `json:"variants"`
}

var homoglyphs = map[rune][]rune{
	'a': {'à', 'á', 'â', 'ã', 'ä', 'å', 'ɑ', 'А', 'Α', 'Ꭺ', 'Ａ', '𝔄', '𝕬', '𝒜', '𝐀', '𝐴', '𝘈', '𝙰', '𝖠', '𝗔', '𝘼', '𝚨', '𝑨', 'ⓐ', 'Ⓐ', '🅐', '🅰', '𝔞', '𝖆', '𝒶', '𝗮', '𝘢', 'ā', 'ă', 'ą', 'ȃ', 'ȧ', 'ạ', 'ả', 'ấ', 'ầ', 'ẩ', 'ẫ', 'ậ', 'ắ', 'ằ', 'ẳ', 'ẵ', 'ặ'},
	'b': {'Ь', 'Ꮟ', 'Ƅ', 'ᖯ', '𝐛', '𝑏', '𝒃', '𝓫', '𝔟', '𝕓', '𝖇', '𝗯', '𝘣', '𝙗', '𝚋', 'ƀ', 'ɓ', 'ḃ', 'ḅ', 'ḇ', 'Ḃ', 'Ḅ', 'Ḇ', 'Ɓ', 'Ƃ', 'ƃ'},
	'c': {'ϲ', 'с', 'ƈ', 'ȼ', 'ḉ', 'ⲥ', '𝐜', '𝑐', '𝒄', '𝓬', '𝔠', '𝕔', '𝖈', '𝗰', '𝘤', '𝙘', '𝚌', 'ć', 'ĉ', 'ċ', 'č', 'ç', 'ḉ', 'ć', 'Ć', 'Ĉ', 'Ċ', 'Č', 'Ç', 'Ḉ', 'Ȼ'},
	'd': {'ԁ', 'ժ', 'Ꮷ', '𝐝', '𝑑', '𝒅', '𝓭', '𝔡', '𝕕', '𝖉', '𝗱', '𝘥', '𝙙', '𝚍', 'ď', 'đ', 'ḋ', 'ḍ', 'ḏ', 'ḑ', 'ḓ', 'Ď', 'Đ', 'Ḋ', 'Ḍ', 'Ḏ', 'Ḑ', 'Ḓ', 'ƌ', 'ɖ', 'ɗ'},
	'e': {'е', 'ҽ', '℮', 'ḛ', 'ḝ', 'ẹ', 'é', 'è', 'ê', 'ë', 'ē', 'ė', 'ę', '𝐞', '𝑒', '𝒆', '𝓮', '𝔢', '𝕖', '𝖊', '𝗲', '𝘦', '𝙚', '𝚎', 'ĕ', 'ę', 'ė', 'ě', 'ȅ', 'ȇ', 'ȩ', 'ḕ', 'ḗ', 'ḙ', 'ḛ', 'ḝ', 'ẽ', 'ẻ', 'ế', 'ề', 'ể', 'ễ', 'ệ'},
	'f': {'ғ', '𝐟', '𝑓', '𝒇', '𝓯', '𝔣', '𝕗', '𝖋', '𝗳', '𝘧', '𝙛', '𝚏', 'ḟ', 'Ḟ', 'ƒ', 'Ƒ', 'ꜰ'},
	'g': {'ɡ', 'ց', '𝐠', '𝑔', '𝒈', '𝓰', '𝔤', '𝕘', '𝖌', '𝗴', '𝘨', '𝙜', '𝚐', 'ĝ', 'ğ', 'ġ', 'ģ', 'ǧ', 'ǵ', 'ḡ', 'Ĝ', 'Ğ', 'Ġ', 'Ģ', 'Ǧ', 'Ǵ', 'Ḡ', 'ƍ', 'ɠ'},
	'h': {'һ', 'հ', 'Ꮒ', 'ℎ', '𝐡', '𝒉', '𝒽', '𝓱', '𝔥', '𝕙', '𝖍', '𝗵', '𝘩', '𝙝', '𝚑', 'ĥ', 'ħ', 'ȟ', 'ḣ', 'ḥ', 'ḧ', 'ḩ', 'ḫ', 'Ĥ', 'Ħ', 'Ȟ', 'Ḣ', 'Ḥ', 'Ḧ', 'Ḩ', 'Ḫ', 'ƕ'},
	'i': {'і', 'ɩ', 'Ꭵ', 'Ⅰ', 'ı', 'í', 'ì', 'î', 'ï', 'ī', 'į', '𝐢', '𝑖', '𝒊', '𝓲', '𝔦', '𝕚', '𝖎', '𝗶', '𝘪', '𝙞', '𝚒', 'ĩ', 'ĭ', 'į', 'ı', 'ǐ', 'ȉ', 'ȋ', 'ḭ', 'ḯ', 'ỉ', 'ị', 'İ', 'Ì', 'Í', 'Î', 'Ï', 'Ĩ', 'Ī', 'Ĭ', 'Į', 'Ǐ', 'Ȉ', 'Ȋ', 'Ḭ', 'Ḯ', 'Ỉ', 'Ị'},
	'j': {'ј', 'ʝ', 'ϳ', '𝐣', '𝑗', '𝒋', '𝓳', '𝔧', '𝕛', '𝖏', '𝗷', '𝘫', '𝙟', '𝚓', 'ĵ', 'ǰ', 'Ĵ', 'ȷ', 'ɉ'},
	'k': {'κ', '𝐤', '𝑘', '𝒌', '𝓴', '𝔨', '𝕜', '𝖐', '𝗸', '𝘬', '𝙠', '𝚔', 'ķ', 'ǩ', 'ḱ', 'ḳ', 'ḵ', 'Ķ', 'Ǩ', 'Ḱ', 'Ḳ', 'Ḵ', 'ƙ', 'ɨ'},
	'l': {'ⅼ', 'ӏ', 'Ɩ', 'ʟ', '𝐥', '𝑙', '𝒍', '𝓵', '𝔩', '𝕝', '𝖑', '𝗹', '𝘭', '𝙡', '𝚕', 'ĺ', 'ļ', 'ľ', 'ŀ', 'ł', 'ḷ', 'ḹ', 'ḻ', 'ḽ', 'Ĺ', 'Ļ', 'Ľ', 'Ŀ', 'Ł', 'Ḷ', 'Ḹ', 'Ḻ', 'Ḽ', 'ƚ', 'ȴ'},
	'm': {'м', 'ṃ', 'ᴍ', '𝐦', '𝑚', '𝒎', '𝓶', '𝔪', '𝕞', '𝖒', '𝗺', '𝘮', '𝙢', '𝚖', 'ḿ', 'ṁ', 'ṁ', 'ṃ', 'Ḿ', 'Ṁ', 'Ṃ', 'ɱ'},
	'n': {'ո', 'п', 'ռ', 'ṅ', 'ṇ', 'ṋ', '𝐧', '𝑛', '𝒏', '𝓷', '𝔫', '𝕟', '𝖓', '𝗻', '𝘯', '𝙣', '𝚗', 'ń', 'ņ', 'ň', 'ǹ', 'ṅ', 'ṇ', 'ṉ', 'ṋ', 'Ń', 'Ņ', 'Ň', 'Ǹ', 'Ṅ', 'Ṇ', 'Ṉ', 'Ṋ', 'ƞ', 'ɲ', 'ŋ'},
	'o': {'ο', 'օ', 'ӧ', 'ö', 'ó', 'ò', 'ô', 'õ', 'ō', 'ő', 'ⲟ', '𝐨', '𝑜', '𝓸', '𝔬', '𝕠', '𝖔', '𝗼', '𝘰', '𝙤', '𝚬', 'ŏ', 'ő', 'ơ', 'ǒ', 'ǫ', 'ǭ', 'ǿ', 'ȍ', 'ȏ', 'ȫ', 'ȭ', 'ȯ', 'ȱ', 'ṍ', 'ṏ', 'ṑ', 'ṓ', 'ọ', 'ỏ', 'ố', 'ồ', 'ổ', 'ỗ', 'ộ', 'ớ', 'ờ', 'ở', 'ỡ', 'ợ'},
	'p': {'р', 'ρ', '⍴', '𝐩', '𝑝', '𝒑', '𝓹', '𝔭', '𝕡', '𝖕', '𝗽', '𝘱', '𝙥', '𝚭', 'ṕ', 'ṗ', 'Ṕ', 'Ṗ', 'ƥ', 'ƿ'},
	'q': {'զ', 'ԛ', 'գ', '𝐪', '𝑞', '𝒒', '𝓺', '𝔮', '𝕢', '𝖖', '𝗾', '𝘲', '𝙦', '𝚞', 'ʠ'},
	'r': {'ᴦ', 'г', 'ř', 'ȓ', 'ṛ', 'ⲅ', '𝐫', '𝑟', '𝒓', '𝓻', '𝔯', '𝕣', '𝖗', '𝗿', '𝘳', '𝙧', '𝚛', 'ŕ', 'ŗ', 'ř', 'ȑ', 'ȓ', 'ṙ', 'ṛ', 'ṝ', 'ṟ', 'Ŕ', 'Ŗ', 'Ř', 'Ȑ', 'Ȓ', 'Ṙ', 'Ṛ', 'Ṝ', 'Ṟ', 'ɍ', 'ɽ', 'ɾ', 'ɿ'},
	's': {'ѕ', 'ʂ', 'ṡ', 'ṣ', '𝐬', '𝑠', '𝒔', '𝓼', '𝔰', '𝕤', '𝖘', '𝘴', '𝙨', '𝚜', 'ś', 'ŝ', 'ş', 'š', 'ș', 'ṡ', 'ṣ', 'ṥ', 'ṧ', 'ṩ', 'Ś', 'Ŝ', 'Ş', 'Š', 'Ș', 'Ṡ', 'Ṣ', 'Ṥ', 'Ṧ', 'Ṩ', 'ƨ', 'ʃ'},
	't': {'т', 'τ', 'ṭ', 'ț', 'ⲧ', '𝐭', '𝑡', '𝒕', '𝓽', '𝔱', '𝕥', '𝖙', '𝘵', '𝙩', '𝚝', 'ţ', 'ť', 'ŧ', 'ț', 'ṫ', 'ṭ', 'ṯ', 'ṱ', 'Ţ', 'Ť', 'Ŧ', 'Ț', 'Ṫ', 'Ṭ', 'Ṯ', 'Ṱ', 'ƚ', 'ƭ', 'ʇ'},
	'u': {'υ', 'ս', 'ü', 'ú', 'ù', 'û', 'ū', 'ⲩ', '𝐮', '𝑢', '𝒖', '𝓾', '𝔲', '𝕦', '𝖚', '𝘶', '𝙪', '𝚞', 'ŭ', 'ů', 'ű', 'ų', 'ư', 'ǔ', 'ǖ', 'ǘ', 'ǚ', 'ǜ', 'ȕ', 'ȗ', 'ṳ', 'ṵ', 'ṷ', 'ṹ', 'ṻ', 'ụ', 'ủ', 'ứ', 'ừ', 'ử', 'ữ', 'ự'},
	'v': {'ν', 'ѵ', 'ⴸ', '𝐯', '𝑣', '𝒗', '𝓿', '𝔳', '𝕧', '𝖛', '𝗏', '𝘷', '𝙫', '𝚟', 'ṽ', 'ṿ', 'Ṽ', 'Ṿ', 'ʋ', 'ʌ'},
	'w': {'ԝ', 'ա', 'ѡ', 'ⲱ', '𝐰', '𝑤', '𝒘', '𝔀', '𝕨', '𝖜', '𝗐', '𝘸', '𝙬', '𝚠', 'ŵ', 'ẁ', 'ẃ', 'ẅ', 'ẇ', 'ẉ', 'ẘ', 'Ŵ', 'Ẁ', 'Ẃ', 'Ẅ', 'Ẇ', 'Ẉ', 'ƺ'},
	'x': {'х', 'ҳ', 'ӿ', '𝐱', '𝑥', '𝒙', '𝔁', '𝕩', '𝖝', '𝗑', '𝘹', '𝙭', '𝚡', 'ẋ', 'ẍ', 'Ẋ', 'Ẍ'},
	'y': {'у', 'ү', 'ӯ', 'ý', 'ÿ', 'ⲩ', '𝐲', '𝑦', '𝒚', '𝔂', '𝕪', '𝖞', '𝗒', '𝘺', '𝙮', '𝚢', 'ŷ', 'ȳ', 'ẏ', 'ỳ', 'ỵ', 'ỷ', 'ỹ', 'Ŷ', 'Ý', 'Ÿ', 'Ȳ', 'Ẏ', 'Ỳ', 'Ỵ', 'Ỷ', 'Ỹ', 'ƴ'},
	'z': {'ᴢ', 'ż', 'ź', 'ž', '𝐳', '𝑧', '𝒛', '𝔃', '𝕫', '𝖟', '𝗓', '𝘻', '𝙯', '𝚣', 'ź', 'ż', 'ž', 'ẑ', 'ẓ', 'ẕ', 'Ź', 'Ż', 'Ž', 'Ẑ', 'Ẓ', 'Ẕ', 'ƶ', 'ȥ', 'ɀ'},
}

func main() {
	var letter = flag.String("l", "", "Single letter to generate variants for")
	var letterFlag = flag.String("letter", "", "Single letter to generate variants for")
	var word = flag.String("w", "", "Word to generate all possible homoglyph combinations for")
	var wordFlag = flag.String("word", "", "Word to generate all possible homoglyph combinations for")
	var stdin = flag.Bool("s", false, "Read input from stdin")
	var stdinFlag = flag.Bool("stdin", false, "Read input from stdin")
	var format = flag.String("f", "simple", "Output format: simple, detailed, json")
	var formatFlag = flag.String("format", "simple", "Output format: simple, detailed, json")
	var maxCombinations = flag.Int("m", 1000, "Maximum number of combinations to generate for words")
	var maxFlag = flag.Int("max", 1000, "Maximum number of combinations to generate for words")
	var help = flag.Bool("h", false, "Show help")
	var helpFlag = flag.Bool("help", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Punycode Homoglyph Generator by @1hehaq\n\n")
		fmt.Fprintf(os.Stderr, "Generate punycode variants of homoglyphs for wordlist generation.\n\n")
		fmt.Fprintf(os.Stderr, "USAGE:\n")
		fmt.Fprintf(os.Stderr, "  %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OPTIONS:\n")
		fmt.Fprintf(os.Stderr, "  -l, --letter string    Single letter to generate variants for\n")
		fmt.Fprintf(os.Stderr, "  -w, --word string      Word to generate all possible homoglyph combinations for\n")
		fmt.Fprintf(os.Stderr, "  -s, --stdin            Read input from stdin\n")
		fmt.Fprintf(os.Stderr, "  -f, --format string    Output format: simple, detailed, json (default \"simple\")\n")
		fmt.Fprintf(os.Stderr, "  -m, --max int          Maximum number of combinations to generate for words (default 1000)\n")
		fmt.Fprintf(os.Stderr, "  -h, --help             Show this help message\n\n")
		// fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
		// fmt.Fprintf(os.Stderr, "  %s -l a                          # Generate variants for letter 'a'\n", os.Args[0])
		// fmt.Fprintf(os.Stderr, "  %s -w test -f detailed           # Generate variants for word 'test' with detailed output\n", os.Args[0])
		// fmt.Fprintf(os.Stderr, "  echo \"hello\" | %s -s            # Read from stdin\n", os.Args[0])
		// fmt.Fprintf(os.Stderr, "  %s -w example -f json            # Generate with JSON format\n", os.Args[0])
		// fmt.Fprintf(os.Stderr, "  %s -w internationalization -m 500 # Limit combinations for large words\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "FORMATS:\n")
		fmt.Fprintf(os.Stderr, "  simple       Just the punycode results (default)\n")
		fmt.Fprintf(os.Stderr, "  detailed     Pretty formatted with headers and counts\n")
		fmt.Fprintf(os.Stderr, "  json         Structured JSON output\n\n")
		fmt.Fprintf(os.Stderr, "SUPPORTED CHARACTERS:\n")
		fmt.Fprintf(os.Stderr, "  a-z (lowercase letters with extensive Unicode homoglyph support)\n")
		fmt.Fprintf(os.Stderr, "  Includes: Cyrillic, Greek, Armenian, Mathematical symbols, Accented characters, Regional variants, and more\n\n")
	}

	flag.Parse()

	if *help || *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	letterInput := getStringFlag(*letter, *letterFlag)
	wordInput := getStringFlag(*word, *wordFlag)
	stdinInput := getBoolFlag(*stdin, *stdinFlag)
	formatInput := getStringFlag(*format, *formatFlag)
	maxInput := getIntFlag(*maxCombinations, *maxFlag)

	flagCount := 0
	if letterInput != "" {
		flagCount++
	}
	if wordInput != "" {
		flagCount++
	}
	if stdinInput {
		flagCount++
	}

	if flagCount == 0 {
		fmt.Fprintf(os.Stderr, "Error: Please provide input via -l, -w, or -s\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if flagCount > 1 {
		fmt.Fprintf(os.Stderr, "Error: Only one input method allowed at a time\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if stdinInput {
		handleStdin(formatInput, maxInput)
	} else if letterInput != "" {
		generateLetterVariants(letterInput, formatInput)
	} else if wordInput != "" {
		generateWordVariants(wordInput, formatInput, maxInput)
	}
}

func getStringFlag(short, long string) string {
	if short != "" {
		return short
	}
	return long
}

func getBoolFlag(short, long bool) bool {
	return short || long
}

func getIntFlag(short, long int) int {
	if short != 1000 {
		return short
	}
	return long
}

func handleStdin(format string, maxCombinations int) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if len(input) == 1 {
			generateLetterVariants(input, format)
		} else if len(input) > 0 {
			generateWordVariants(input, format, maxCombinations)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

func generateLetterVariants(letter string, format string) {
	if len(letter) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Empty letter provided\n")
		return
	}

	letterRune := unicode.ToLower(rune(letter[0]))
	variants, exists := homoglyphs[letterRune]
	if !exists {
		fmt.Fprintf(os.Stderr, "No homoglyphs found for letter: %c\n", letterRune)
		return
	}

	switch format {
	case "json":
		output := LetterOutput{
			Letter:   string(letterRune),
			Variants: make([]Variant, 0, len(variants)),
		}
		for _, variant := range variants {
			punycode := encodePunycode(string(variant))
			output.Variants = append(output.Variants, Variant{
				Glyph:    string(variant),
				Punycode: punycode,
			})
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	case "detailed":
		fmt.Printf("Punycode variants for letter: '%c'\n", letterRune)
		for i := 0; i < 40; i++ {
			fmt.Print("=")
		}
		fmt.Println()
		for _, variant := range variants {
			punycode := encodePunycode(string(variant))
			if punycode != "" {
				fmt.Printf("%c -> %s\n", variant, punycode)
			} else {
				fmt.Printf("%c -> [ENCODING FAILED]\n", variant)
			}
		}
		for i := 0; i < 40; i++ {
			fmt.Print("=")
		}
		fmt.Println()
		fmt.Printf("Total variants: %d\n", len(variants))
	default:
		for _, variant := range variants {
			punycode := encodePunycode(string(variant))
			if punycode != "" {
				fmt.Println(punycode)
			}
		}
	}
}

func generateWordVariants(word string, format string, maxCombinations int) {
	wordLower := strings.ToLower(word)
	chars := []rune(wordLower)
	
	totalCombinations := 1
	for _, char := range chars {
		if variants, exists := homoglyphs[char]; exists {
			totalCombinations *= (len(variants) + 1)
		}
	}
	
	if totalCombinations > maxCombinations {
		fmt.Fprintf(os.Stderr, "Warning: %d total combinations possible, limiting to %d\n", 
			totalCombinations, maxCombinations)
	}
	
	var results []string
	count := 0
	firstResult := true
	
	switch format {
	case "json":
		fmt.Println("{")
		fmt.Printf("  \"word\": \"%s\",\n", word)
		fmt.Println("  \"variants\": [")
	case "detailed":
		fmt.Printf("Punycode variants for word: '%s'\n", word)
		for i := 0; i < 40; i++ {
			fmt.Print("=")
		}
		fmt.Println()
		fmt.Printf("Generating up to %d combinations from %d total possible\n", 
			maxCombinations, totalCombinations)
		for i := 0; i < 40; i++ {
			fmt.Print("=")
		}
		fmt.Println()
	}
	
	generateCombinations(chars, 0, "", format, maxCombinations, &count, &results, &firstResult)

	switch format {
	case "json":
		fmt.Println("  ],")
		fmt.Printf("  \"total\": %d\n}\n", count)
	case "detailed":
		for i := 0; i < 40; i++ {
			fmt.Print("=")
		}
		fmt.Println()
		fmt.Printf("Total combinations: %d\n", count)
	}
}

func generateCombinations(chars []rune, pos int, current string, format string, maxCombinations int, count *int, results *[]string, firstResult *bool) {
	if *count >= maxCombinations {
		return
	}

	if pos == len(chars) {
		punycode := encodePunycode(current)
		if punycode != "" {
			*count++
			switch format {
			case "json":
				if !*firstResult {
					fmt.Print(",\n")
				}
				*firstResult = false
				fmt.Printf("    \"%s\"", punycode)
			case "detailed":
				fmt.Printf("%s -> %s\n", current, punycode)
			default:
				fmt.Println(punycode)
			}
			*results = append(*results, punycode)
		}
		return
	}

	char := chars[pos]
	generateCombinations(chars, pos+1, current+string(char), format, maxCombinations, count, results, firstResult)

	if variants, exists := homoglyphs[char]; exists {
		for _, variant := range variants {
			generateCombinations(chars, pos+1, current+string(variant), format, maxCombinations, count, results, firstResult)
		}
	}
}

func encodePunycode(s string) string {
	encoded, err := idna.ToASCII(s)
	if err != nil {
		return ""
	}
	return encoded
}
