package fonts

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Name
		wantErr bool
	}{
		{name: "anton", input: "anton", want: Anton},
		{name: "archivo black", input: "archivo-black", want: ArchivoBlack},
		{name: "bebas neue", input: "bebas-neue", want: BebasNeue},
		{name: "inter", input: "inter", want: Inter},
		{name: "unknown", input: "comic-sans", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFile(t *testing.T) {
	for _, name := range Names() {
		t.Run(string(name), func(t *testing.T) {
			data, err := File(name)
			if err != nil {
				t.Fatalf("File(%q) error = %v", name, err)
			}
			if len(data) == 0 {
				t.Errorf("File(%q) returned no data", name)
			}
		})
	}
}
