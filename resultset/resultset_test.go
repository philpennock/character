package resultset

import (
	"errors"
	// make the current module available under same namespace callers would use:
	resultset "."

	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/liquidgecka/testlib"

	"github.com/philpennock/character/sources"
	//"github.com/philpennock/character/table"
	"github.com/philpennock/character/unicode"
)

func Test000LoadDataExternal(*testing.T) {
	// just load data once so that times for other tests are sane
	_ = sources.NewAll()
}

func TestHaveTableSupport(t *testing.T) {
	if !CanTable() {
		t.Error("missing Table support")
	}
}

func TestResultSetBasics(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()

	srcs := sources.NewAll()
	rs := resultset.New(srcs, 5)

	T.Equal(rs.ErrorCount(), 0, "no errors in fresh resultset")

	ci := unicode.CharInfo{
		Number: '✓',
		Name:   "CHECK MARK",
	}

	realStdout := os.Stdout
	defer func() { os.Stdout = realStdout }()
	f, err := ioutil.TempFile("", "testStdout")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { os.Remove(f.Name()) }()
	os.Stdout = f
	getContents := func() ([]byte, error) {
		contents := make([]byte, 2048)
		l, err := f.ReadAt(contents, 0)
		if err != nil && err == io.EOF {
			return contents[:l], nil
		}
		return contents[:l], err
	}

	shouldBe := func(desired, msg string, reset bool) {
		if reset {
			T.ExpectSuccess(f.Truncate(0), "truncating existing output tempfile")
			_, err := f.Seek(0, 0)
			T.ExpectSuccess(err, "seeking existing output tempfile")
		}
		rs.PrintPlain(resultset.PRINT_RUNE)
		c, err := getContents()
		T.ExpectSuccess(err, "getting contents of temp file")
		T.Equal(string(c), desired, msg)
	}

	rs.AddCharInfo(ci)
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n", "printed two check marks", true)
	shouldBe("✓\n✓\n✓\n✓\n", "printed another two check marks", false)
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n✓\n", "printed three check marks", true)
	rs.AddDivider()
	shouldBe("✓\n✓\n✓\n\n", "printed three check marks and divider", true)
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n✓\n\n✓\n", "printed some check marks with divider in there", true)

	rs.AddError("dummy", errors.New("pseudo-error goes here"))
	rs.AddCharInfo(ci)
	realStderr := os.Stderr
	os.Stderr = f
	defer func() { os.Stderr = realStderr }()
	shouldBe("✓\n✓\n✓\n\n✓\nlooking up \"dummy\": pseudo-error goes here\n✓\n", "printed some check-marks and an error", true)
}