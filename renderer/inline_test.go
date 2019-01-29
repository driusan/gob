package renderer

import (
	"context"
	"image"
	"testing"
)

// Tests that an inline with an explicit line-height that doesn't match the line
// height gets its line advanced according to the line-height, and positioned
// according to the half-leading.
func TestInlineLineHeight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="line-height: 20px; font-size: 14px;">
			This is inline text within the body. It has a line-height of 20px,
and a font-size of 14px. This means that each line should be exactly 20px below
the previous one, with the 6px evenly distributed above and below the line.
To test this. We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	// "lineBox" 0 is the first character, so we need to be sure that
	// there's 3 boxes to ensure multiple lines.
	if len(page.Content.lineBoxes) < 3 {
		t.Fatal("Not enough lines of text to test line height.")
	}

	for i, line := range page.Content.lineBoxes[1:] {
		if want := i*20 + 3; line.origin.Y != want {
			t.Errorf("Line %d at incorrect height. got %v want %v", i, line.origin.Y, want)
		}
	}
}

// Tests that a line-height that's smaller than the font-size gets adjusted with
// a negative half-leading.
func TestInlineSmallLineHeight(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="line-height: 14px; font-size: 20px;">
			This is inline text within the body. It has a line-height of 14px,
and a font-size of 20px. This means that each line should be exactly 14px below
the previous one, with the -6px evenly distributed above and below the line causing
the text to overlap. To test this. We need to ensure that there are multiple lines. This should be
enough text.
		</body>
	</html>`,
	)

	size := image.Point{400, 300}
	page.Content.Layout(context.TODO(), size)

	// "lineBox" 0 is the first character, so we need to be sure that
	// there's 3 boxes to ensure multiple lines.
	if len(page.Content.lineBoxes) < 3 {
		t.Fatal("Not enough lines of text to test line height.")
	}

	for i, line := range page.Content.lineBoxes[1:] {
		if want := i*14 - 3; line.origin.Y != want {
			t.Errorf("Line %d at incorrect height. got %v want %v", i, line.origin.Y, want)
		}
	}
}

// Test that borders on inline elements which span multiple lines are positioned
// appropriately.
func TestInlineBorder(t *testing.T) {
	page := parseHTML(
		t,
		`<html>
		<body style="line-height: 20px; font-size: 12px">
		before <span style="border: 1px solid green; padding: 1px;">
			This is inline text within the body. It has a line-height of 20px,
and a font-size of 12px. This means that each line should be exactly 20px below
the previous one, with the 8px evenly distributed above and below the line. There is
1px of padding and a 1px green border around the text. but the left and right border
should only be on the first and last line, respectively.</span>
		</body>
	</html>`,
	)

	canvas := image.NewRGBA(image.Rectangle{image.ZP, image.Point{400, 300}})
	size := canvas.Bounds().Size()
	page.Content.Layout(context.TODO(), size)
	page.Content.RenderInto(context.TODO(), canvas, image.ZP)

	// count the number of green pixels in a row in canvas.
	greenPx := func(row int) int {
		rv := 0
		for i := 0; i < size.X; i++ {
			r, g, b, a := canvas.At(i, row).RGBA()
			if r == 0 && g > 32000 && b == 0 && a == 65535 {
				rv++
			}
		}
		return rv
	}

	// "lineBox" 0 is the first character, lineBox 1 is "efore", and lineBox
	// 2 is the first letter of the span. So we need to be sure that
	// there's 6 boxes to ensure 3 lines (one with no left or right border,
	// one with a left border, and one with a right border.)
	if len(page.Content.lineBoxes) < 6 {
		t.Fatal("Not enough lines of text to test inline borders.")
	}

	lines := page.Content.lineBoxes[2:]
	for i, line := range lines {
		if want := i*20 + 4; line.origin.Y != want {
			t.Errorf("Line %d at incorrect height. got %v want %v", i, line.origin.Y, want)
		}

		switch i {
		case 0, len(lines) - 1:
			// For the first and last line, the top border, we
			// have a smaller threshhold for how many pixels need
			// to exist for the top and bottom border and only check
			// that they're equal, since we don't know exactly where
			// the line stops and ends.
			topgreen := greenPx(line.origin.Y - 2)
			if topgreen < 100 {
				t.Errorf("Line %d top border missing got %v green pixels want >= 100", i, topgreen)
			}

			// origin, plus 2 padding (one for top and one for bottom),
			// plus the font size, then the next row is the border itself
			bottomgreen := greenPx(line.origin.Y + 12 + 2 + 1)
			if bottomgreen != topgreen {
				t.Errorf("Line %d bottom border does not match top border. got %v green pixels want >= %v", i, bottomgreen, topgreen)
			}

			// We don't know exactly where the side border is, we only know
			// that it's somewhere in the line and there should only
			// be one.
			if middlegreen := greenPx(line.origin.Y + 12); middlegreen != 1 {
				t.Errorf("Line %d expected only 1px of green border in middle of first line and last line. got %v", i, middlegreen)
			}
		default:
			// Middle lines. We don't know exactly where the cutoff
			// of the line is, but most of it should be green.
			topgreen := greenPx(line.origin.Y - 2)
			if topgreen < 350 {
				t.Errorf("Line %d missing top border. got %v green pixels want ~400", i, topgreen)
			}
			// 12 for the font size, 2 for the padding, and then the
			// next one actually has the border.
			bottomgreen := greenPx(line.origin.Y + 12 + 2 + 1)
			if bottomgreen != topgreen {
				t.Errorf("Line %d bottom border does not match top border. got %v green pixels want >= %v", i, bottomgreen, topgreen)
			}

			if middlegreen := greenPx(line.origin.Y + 12); middlegreen != 0 {
				t.Errorf("Line %d unexpected left or right border. got %v green pixels", i, middlegreen)
			}

		}
	}

}

