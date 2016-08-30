# from : http://www.chris.com/ascii/index.php?art=animals/elephants
cat etc/refactor/progress.txt

total=`grep -R "func Test" ./src/* | grep -v vendor | grep -v "func TestMain" | wc -l`
passing=`grep -R "func Test.*RF" ./src/* | grep -v vendor | grep -v "func TestMain" | wc -l`

echo 'package main; import "fmt"; import "os"; import "strconv";
func main() { a, _ := strconv.ParseFloat(os.Args[1], 64); b, _ := strconv.ParseFloat(os.Args[2], 64); fmt.Printf("%0.2f\n", 100.0*a/b); }' > /tmp/progress.go

percentage=$(go run /tmp/progress.go $passing $total)


echo "Pete the Pachyderm says ... $percentage% way done!"
echo "Only $(($total - $passing)) more tests to go!"

# Print remaining tests:
echo "Remaining tests:"
grep -R "func Test" ./src | grep -v vendor | grep -v "func TestMain" | grep -v "RF("  | cut -f 2 -d " " | cut -f 1 -d "("

