package app

import (
	"os"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/dsoprea/go-jpeg-image-structure/v2"
)

// updateNativeExif updates the EXIF data of a JPEG file.
func updateNativeExif(path string, lat, lon float64, dateTime time.Time) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(data)
	if err != nil {
		return err
	}

	sl := intfc.(*jpegstructure.SegmentList)
	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		return err
	}

	ifdIb, _ := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0")
	exifIb, _ := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0/Exif")
	gpsIb, _ := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0/GPSInfo")

	dtStr := dateTime.Format("2006:01:02 15:04:05")
	_ = ifdIb.SetStandardWithName("DateTime", dtStr)
	_ = exifIb.SetStandardWithName("DateTimeOriginal", dtStr)
	_ = exifIb.SetStandardWithName("CreateDate", dtStr)

	if lat != 0 || lon != 0 {
		latRef, lonRef := "N", "E"
		if lat < 0 {
			latRef, lat = "S", -lat
		}
		if lon < 0 {
			lonRef, lon = "W", -lon
		}

		_ = gpsIb.SetStandardWithName("GPSLatitudeRef", latRef)
		_ = gpsIb.SetStandardWithName("GPSLongitudeRef", lonRef)
		_ = gpsIb.SetStandardWithName("GPSLatitude", decimalToRationals(lat))
		_ = gpsIb.SetStandardWithName("GPSLongitude", decimalToRationals(lon))
	}

	_ = sl.SetExif(rootIb)
	f, _ := os.Create(path)
	defer f.Close()
	_ = sl.Write(f)
	return os.Chtimes(path, dateTime, dateTime)
}

// decimalToRationals converts a decimal degree to EXIF rational format.
func decimalToRationals(decimal float64) []exifcommon.Rational {
	degrees := int(decimal)
	minutesFloat := (decimal - float64(degrees)) * 60
	minutes := int(minutesFloat)
	seconds := (minutesFloat - float64(minutes)) * 60
	return []exifcommon.Rational{
		{Numerator: uint32(degrees), Denominator: 1},
		{Numerator: uint32(minutes), Denominator: 1},
		{Numerator: uint32(seconds * 1000), Denominator: 1000},
	}
}
