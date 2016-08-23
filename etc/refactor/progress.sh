# from : http://www.chris.com/ascii/index.php?art=animals/elephants
cat etc/refactor/progress.txt

total=0
pfs_total=`grep -R "func Test" ./src/server/pfs/* | grep -v vendor | grep -v "func TestMain" | wc -l`
total=$(($total + $pfs_total))

integration_total=`grep -R "func Test" ./src/server/pachyderm_test.go | grep -v vendor | grep -v "func TestMain" | wc -l`
total=$(($total + $integration_total))

passing=0
pfs_passing=`grep -R "func Test.*RF" ./src/server/pfs/* | grep -v vendor | grep -v "func TestMain" | wc -l`
passing=$(($passing + $pfs_passing))
integration_passing=`grep -R "func Test.*RF" ./src/server/pachyderm_test.go | grep -v vendor | grep -v "func TestMain" | wc -l`
passing=$(($passing + $integration_passing))

#echo "$total total"
#echo "$passing passing"

echo 'package main; import "fmt"; import "os"; import "strconv";
func main() { a, _ := strconv.ParseFloat(os.Args[1], 64); b, _ := strconv.ParseFloat(os.Args[2], 64); fmt.Printf("%0.2f\n", 100.0*a/b); }' > /tmp/progress.go

percentage=$(go run /tmp/progress.go $passing $total)


echo "Pete the Pachyderm says ... $percentage% way done!"
echo "Only $(($total - $passing)) more tests to go!"
