package util

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
	"go.uber.org/zap/zapcore"
)

type DateDiffResponse struct {
	Year, Month, Day, Hour, Min, Sec int
}

// Helpers is a struct that contains the compiled regexes
type Helpers struct {
	URLRegex *regexp.Regexp
}

// ParseURL removes protocol and www. from URls
// TODO: Add example
func (helpers *Helpers) ParseURL(url string) string {
	if len(url) > 0 {
		cleanedURL := helpers.URLRegex.ReplaceAllString(strings.TrimRight(url, "/"), "")
		return strings.Trim(cleanedURL, " ")
	}

	return ""
}

func (helpers *Helpers) DeSanitizePlain(text string) string {
	p := bluemonday.StrictPolicy()
	result := p.Sanitize(text)
	result = strings.ReplaceAll(result, "\\", "")
	result = html.UnescapeString(result)
	// result = strings.Trim(result, ".\"")
	result = strings.ReplaceAll(result, "&#44;", "")
	return strings.Trim(result, " ")
}

// Init initializes the init variables
func (helpers *Helpers) Init() {
	helpers.URLRegex = regexp.MustCompile(`https?:\/?\/|www.`)
}

func DateDiff(a, b time.Time) DateDiffResponse {

	diff := DateDiffResponse{}

	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}

	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	diff.Year = int(y2 - y1)
	diff.Month = int(M2 - M1)
	diff.Day = int(d2 - d1)
	diff.Hour = int(h2 - h1)
	diff.Min = int(m2 - m1)
	diff.Sec = int(s2 - s1)

	// Normalize negative values
	if diff.Sec < 0 {
		diff.Sec += 60
		diff.Min--
	}
	if diff.Min < 0 {
		diff.Min += 60
		diff.Hour--
	}
	if diff.Hour < 0 {
		diff.Hour += 24
		diff.Day--
	}
	if diff.Day < 0 {
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		diff.Day += 32 - t.Day()
		diff.Month--
	}
	if diff.Month < 0 {
		diff.Month += 12
		diff.Year--
	}

	return diff
}

func ArrayDiff(a, b []string) []string {
	a = sortIfNeeded(a)
	b = sortIfNeeded(b)
	var d []string
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		c := strings.Compare(a[i], b[j])
		if c == 0 {
			i++
			j++
		} else if c < 0 {
			d = append(d, a[i])
			i++
		} else {
			d = append(d, b[j])
			j++
		}
	}
	d = append(d, a[i:len(a)]...)
	d = append(d, b[j:len(b)]...)
	return d
}

func Shuffle(vals []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(vals) > 0 {
		n := len(vals)
		randIndex := r.Intn(n)

		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
		vals = vals[:n-1]
	}
}

func sortIfNeeded(a []string) []string {
	if sort.StringsAreSorted(a) {
		return a
	}
	s := append(a[:0:0], a...)
	sort.Strings(s)
	return s
}

func ArrayChunk(s []interface{}, size int) [][]interface{} {
	if size < 1 {
		panic("size: cannot be less than 1")
	}

	length := len(s)
	chunks := int(math.Ceil(float64(length) / float64(size)))
	var n [][]interface{}
	for i, end := 0, 0; chunks > 0; chunks-- {
		end = (i + 1) * size
		if end > length {
			end = length
		}
		n = append(n, s[i*size:end])
		i++
	}
	return n
}

func ArrayKeys(elements interface{}) []interface{} {
	m := elements.(map[interface{}]interface{})

	i, keys := 0, make([]interface{}, len(m))
	for key := range m {
		keys[i] = key
		i++
	}
	return keys
}

func MergeMap(a map[string]string, b map[string]string) {
	for k, v := range b {
		a[k] = v
	}
}

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func FormatNumToSI(num int) string {
	SISymbol := []string{"", "k", "M", "B"}

	// what tier? (determines SI symbol)
	var tier = int(math.Log10(float64(num)) / 3)
	var scale = math.Pow(10, float64(tier*3))

	// scale the number
	var scaled = strconv.FormatFloat(float64(num)/scale, 'f', 1, 64)

	// Handle 10.0k case
	if strings.Index(scaled, ".0") > -1 {
		scaled = strings.ReplaceAll(scaled, ".0", "")
	}

	if tier >= 0 && tier < len(SISymbol) {
		// format number and add suffix
		return scaled + SISymbol[tier]
	}

	return strconv.Itoa(num)
}

// GetBaseURL returns
/*
* Returns the base url of the link provided WITHOUT the Scheme (e.g. the protocol http or https), if successfully parsed.
* Otherwise, returns an empty string
*
*
* Input: https://google.com/blehbelh/leh/blezds/adasd/asdas/d?q=adsd&s=asds
* Output: google.com
*
*
 */
func GetBaseURL(link string, withScheme bool) string {
	var baseURL string
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}

	if withScheme {
		baseURL += u.Scheme + "://"
	}

	baseURL += u.Hostname()

	return baseURL
}

func pkcs7pad(buf []byte, size int) []byte {
	if size < 1 || size > 255 {
		panic(fmt.Sprintf("pkcs7pad: inappropriate block size %d", size))
	}
	i := size - (len(buf) % size)
	return append(buf, bytes.Repeat([]byte{byte(i)}, i)...)
}

// Returns slice of the original data without padding.
func pkcs7Unpad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	if len(data)%blocklen != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data len %d", len(data))
	}
	padlen := int(data[len(data)-1])
	if padlen > blocklen || padlen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	// check padding
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if pad[i] != byte(padlen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padlen], nil
}

func Decrypt(key []byte, iv []byte, encrypted string) ([]byte, error) {

	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)
	h.Reset()
	h.Write(iv)
	iv = h.Sum(nil)[:16]

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("bad blocksize(%v), aes.BlockSize = %v\n", len(data), aes.BlockSize)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(c, iv)
	cbc.CryptBlocks(data, data)
	out, err := pkcs7Unpad(data, aes.BlockSize)
	if err != nil {
		return out, err
	}
	return out, nil
}

func Encrypt(key []byte, iv []byte, plaintext []byte) string {

	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)
	h.Reset()
	h.Write(iv)
	iv = h.Sum(nil)[:16]

	plainTextPadded := pkcs7pad(plaintext, aes.BlockSize)

	if len(plainTextPadded)%aes.BlockSize != 0 {
		panic("plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(plainTextPadded, plainTextPadded)

	return base64.StdEncoding.EncodeToString(plainTextPadded)
}

func JoinInts(collection []int, sep string) string {
	concatenated := ""

	for _, j := range collection {
		concatenated += strconv.Itoa(j)
		concatenated += ","
	}

	concatenated = concatenated[0 : len(concatenated)-1]

	return concatenated
}

func CapFirstChar(s string) string {
	for index, value := range s {
		return string(unicode.ToUpper(value)) + s[index+1:]
	}
	return ""
}

func FloatToInt(num float64) int {
	numString := strconv.FormatFloat(num, 'f', 0, 64)
	numInt, _ := strconv.Atoi(numString)
	return numInt
}

func ParseAmbiguousFloatString(num interface{}, precision int) float64 {
	switch num.(type) {
	case string:
		{
			numFloat, _ := strconv.ParseFloat(num.(string), 64)
			return ToFixed(numFloat, precision)
		}
	case float64:
		{
			return ToFixed(num.(float64), precision)
		}
	}

	return 0.0
}

func ParseAmbiguousNullString(text interface{}) string {
	switch text.(type) {
	case string:
		{
			return text.(string)
		}
	case float64:
		{
			strconv.FormatFloat(text.(float64), 'f', 0, 64)
		}
	case nil:
		{
			return ""
		}
	}

	return ""
}

func ParseAmbiguousNullStringArray(value interface{}) []string {
	switch value.(type) {
	case []interface{}:
		{
			result := make([]string, 0)
			for _, v := range value.([]interface{}) {
				result = append(result, v.(string))
			}
			return result
		}
	case []string:
		{
			return value.([]string)
		}
	case nil:
		{
			return []string{}
		}
	}

	return []string{}
}

func ParseAmbiguousNullInt(value interface{}) int {
	switch value.(type) {
	case int:
		{
			return value.(int)
		}
	case string:
		{
			return FloatToInt(ParseAmbiguousFloatString(value, 0))

		}
	case float64:
		{
			return FloatToInt(value.(float64))
		}
	case nil:
		{
			return 0
		}
	}

	return 0
}

func ParseAmbiguousNullBool(value interface{}) bool {
	switch value.(type) {
	case int:
		{
			return value.(int) > 0
		}
	case string:
		{
			b, _ := strconv.ParseBool(value.(string))
			return b
		}
	case nil:
		{
			return false
		}
	}

	return false

}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(math.Round(num*output)) / output
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func FormatDate(dateText string, inputFormat string, outputFormat string) string {

	t, parseErr := time.Parse(inputFormat, dateText)

	if parseErr != nil {
		return ""
	}

	return t.Format(outputFormat)
}

type Fasty struct {
	Mux sync.Mutex
	Wg  sync.WaitGroup
}

func UnescapeUnicodeCharactersInJSON(_jsonRaw json.RawMessage) (json.RawMessage, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(_jsonRaw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func Stringify(data interface{}) string {
	byteData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("Error stringifying. Err = %v", err)
	}

	return string(byteData)
}

func PrettyStringify(data interface{}) string {
	byteData, err := json.MarshalIndent(data, "  ", "    ")
	if err != nil {
		return fmt.Sprintf("Error stringifying. Err = %v", err)
	}

	return string(byteData)
}

func IsCodeValid(code int) bool {
	return code == 200
}

func MakeDir(path string, mode os.FileMode) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, mode)
	}
}

func GetLogLevel() zapcore.Level {
	logLevel := os.Getenv("LOG_LEVEL")

	zapLogLevel := zapcore.DebugLevel
	if logLevel != "debug" {
		zapLogLevel = zapcore.InfoLevel
	}

	return zapLogLevel
}

func LoadEnv(env string) {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	ap := path.Join(basepath, "../../", env)
	if err := godotenv.Load(ap); err != nil {
		log.Fatalf("%s", err)
	}
}

func CopyOutputFromIOReader(output *string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		*output += scanner.Text()
	}
}

func ArrayIntersect(a, b []string) []string {
	freqArr := map[string]int{}

	var commonElems []string
	for _, element := range a {
		freqArr[element]++
	}

	for _, element := range b {
		if freqArr[element] > 0 {
			commonElems = append(commonElems, element)
			freqArr[element]--
		}
	}

	return commonElems
}

var (
	dictionary = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
)

// Encode converts the big integer to alpha id (an alphanumeric id with mixed cases)
func AlphaEncode(val string) string {
	var result []byte
	var index int
	var strVal string

	base := big.NewInt(int64(len(dictionary)))
	a := big.NewInt(0)
	b := big.NewInt(0)
	c := big.NewInt(0)
	d := big.NewInt(0)

	exponent := 1

	remaining := big.NewInt(0)
	remaining.SetString(val, 10)

	for remaining.Cmp(big.NewInt(0)) != 0 {
		a.Exp(base, big.NewInt(int64(exponent)), nil) //16^1 = 16
		b = b.Mod(remaining, a)                       //119 % 16 = 7 | 112 % 256 = 112
		c = c.Exp(base, big.NewInt(int64(exponent-1)), nil)
		d = d.Div(b, c)

		//if d > dictionary.length, we have a problem. but BigInteger doesnt have
		//a greater than method :-(  hope for the best. theoretically, d is always
		//an index of the dictionary!
		strVal = d.String()
		index, _ = strconv.Atoi(strVal)
		result = append(result, dictionary[index])
		remaining = remaining.Sub(remaining, b) //119 - 7 = 112 | 112 - 112 = 0
		exponent = exponent + 1
	}

	//need to reverse it, since the start of the list contains the least significant values
	return string(reverse(result))
}

// Decode converts the alpha id to big integer
func AlphaDecode(s string) string {
	//reverse it, coz its already reversed!
	chars2 := reverse([]byte(s))

	//for efficiency, make a map
	dictMap := make(map[byte]*big.Int)

	j := 0
	for _, val := range dictionary {
		dictMap[val] = big.NewInt(int64(j))
		j = j + 1
	}

	bi := big.NewInt(0)
	base := big.NewInt(int64(len(dictionary)))

	exponent := 0
	a := big.NewInt(0)
	b := big.NewInt(0)
	intermed := big.NewInt(0)

	for _, c := range chars2 {
		a = dictMap[c]
		intermed = intermed.Exp(base, big.NewInt(int64(exponent)), nil)
		b = b.Mul(intermed, a)
		bi = bi.Add(bi, b)
		exponent = exponent + 1
	}
	return bi.String()
}

func reverse(bs []byte) []byte {
	for i, j := 0, len(bs)-1; i < j; i, j = i+1, j-1 {
		bs[i], bs[j] = bs[j], bs[i]
	}
	return bs
}

func DecryptWithRandomIV(key []byte, encrypted string) ([]byte, error) {

	data, err := hex.DecodeString(encrypted)

	if err != nil {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("bad blocksize(%v), aes.BlockSize = %v\n", len(ciphertext), aes.BlockSize)
	}

	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cbc := cipher.NewCBCDecrypter(c, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)
	out, err := pkcs7Unpad(ciphertext, aes.BlockSize)
	if err != nil {
		return out, err
	}
	return out, nil
}

func EncryptWithRandomIV(key []byte, plaintext []byte) string {

	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)
	h.Reset()

	plainTextPadded := pkcs7pad(plaintext, aes.BlockSize)

	if len(plainTextPadded)%aes.BlockSize != 0 {
		return ""
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	cipherText := make([]byte, aes.BlockSize+len(plainTextPadded))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(cryptoRand.Reader, iv); err != nil {
		return ""
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainTextPadded)

	return hex.EncodeToString(cipherText)
}
