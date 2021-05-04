package image

import (
	"io"

	"gitlab.com/gomidi/muskel/score"
	"gitlab.com/gomidi/smfimage"
)

/*

func Background(name string) Option {
		case "black":
		case "white":
		case "transparent":
func BeatsInGrid() Option {
func Curve() Option {
func SingleBar() Option {
func Monochrome() Option {
func NoBackground() Option {
func NoBarLines() Option {

// BaseNote sets the reference/base note
// smfimage.C , smfimage.CSharp etc.
func BaseNote(n Note) Option {

// Height of 32thnote in pixel (default = 4)
func Height(height int) Option {

// Width of 32thnote in pixel (default = 4)
func Width(width int) Option {

// TrackBorder in pixel (default = 2)
func TrackBorder(border int) Option {

func TrackOrder(order ...int) Option {
func SkipTracks(tracks ...int) Option {
func Colors(cm ColorMapper) Option {
func Overview() Option {
func Verbose() Option {
*/

func MakeImage(sc *score.Score, smfInput io.Reader, outFile string, opts ...smfimage.Option) error {
	options, err := sc.GetOptionsForImage()
	if err != nil {
		return err
	}
	options = append(options, opts...)

	options = append(options, smfimage.BeatsInGrid(), smfimage.BaseNote(smfimage.C))

	smfimg, _, err := smfimage.New(smfInput, options...)
	if err != nil {
		return err
	}
	return smfimg.SaveAsPNG(outFile)
}
