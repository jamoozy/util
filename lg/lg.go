// A more natural log than "log".
package lg

import (
  "flag"
  "errors"
  "fmt"
  "io"
  "log"
  "os"
  "strconv"
)


////////////////////////////////////////////////////////////////////////////////
//                         Global Verbosity Variable                          //
////////////////////////////////////////////////////////////////////////////////

// The verbosity var; meant only to be used with flag.Var().
var vv = verbVar{0}

// Sets up the command line arguments for logging.
func init() {
  flag.Var(&vv, "verbosity", "Set verbosity in [-3,3].")
}

// ---- Actual flag definition (using Value interface).

// Custom verbosity variable.
type verbVar struct {
  v int8
}

// Shows verbosity.
func (v *verbVar) String() string {
  return fmt.Sprintf("%d", v.v)
}

// Ups the verbosity
func (v *verbVar) Set(value string) error {
  pVal, err := strconv.ParseInt(value, 10, 64)
  if err == nil {
    v.v = int8(pVal)
    return Set(v.v)
  }
  return err
}



////////////////////////////////////////////////////////////////////////////////
//                                   Logger                                   //
////////////////////////////////////////////////////////////////////////////////

// A lg-type Logger.  Is en- or disabled based on verbosity level in VerbVar.
type Logger struct {
  *log.Logger
  activeLevel int8    // The level at which this activates.
  enabled bool        // Whether this will actually log.
}

// Expanded logging util.  `level` is the level at or above which this logger
// will log.  `enabled` is whether to enable the logger.
func New(out io.Writer, prefix string, flag, level int, enabled bool) *Logger {
  return &Logger{log.New(out, prefix, flag), int8(level), enabled, ""}
}

// En- or disables the Logger.
func (l *Logger) Enable(enable bool) {
  l.enabled = enable
}

// Overwrite default log behavior to include the indent string.
func (l *Logger) Print(v ...interface{}) {
  if l.enabled && l.activeLevel >= level {
    l.Output(calldepth, indent + fmt.Sprint(v))
  }
}
func (l *Logger) Printf(format string, v ...interface{}) {
  if l.enabled && l.activeLevel >= level {
    l.Output(calldepth, indent + fmt.Sprintf(format, v...))
  }
}
func (l *Logger) Println(v ...interface{}) {
  if l.enabled && l.activeLevel >= level {
    l.Output(calldepth, indent + fmt.Sprintln(v...))
  }
}



////////////////////////////////////////////////////////////////////////////////
//                      Global Logging Functions & Vars                       //
////////////////////////////////////////////////////////////////////////////////

// Printing at various levels of urgency.  These loggers log to their respective
// locations (or not) based on the verbosity level that was set.
var (
  TrcLg = New(os.Stdout, "[trc] ", 0,  3, false)   // Trace logger.
  DbgLg = New(os.Stdout, "[dbg] ", 0,  2, false)   // Debug logger.
  VrbLg = New(os.Stdout, "[vrb] ", 0,  1, false)   // Verbose logger.
  IfoLg = New(os.Stdout, "[ifo] ", 0,  0,  true)   // Info logger.
  WrnLg = New(os.Stdout, "[wrn] ", 0, -1,  true)   // Warn logger.
  ErrLg = New(os.Stderr, "[err] ", 0, -2,  true)   // Error logger.
  FtlLg = New(os.Stderr, "[ftl] ", 0, -3,  true)   // Fatal logger.
)

// Convenience variable used for looping through all the loggers to perform the
// same task on all of them.
var lgs = []*Logger{TrcLg, DbgLg, VrbLg, IfoLg, WrnLg, ErrLg, FtlLg}

// Log level we're currently at.
var level int8

// Variables used for indenting.
var (
  indent string   // The actual text of the indentation (usually jut a bunch of
                  // spaces).
  calldepth int   // The number of indents.
)

// Sets the verbosity level.  If v is too large or too small, sets verbosity to
// the closest level, and returns an error.
func Set(v int8) (err error) {
  // Verify a good range.
  if v < -3 {
    err = errors.New(fmt.Sprintf("%d , min:-3", v))
  } else if 3 < v {
    err = errors.New(fmt.Sprintf("%d > max:3", v))
  }

  // Set the level regardless of error state.
  level = v

  // Hopefully return nil.
  return err
}

// Convenience, so you don't have to do, e.g., lg.TrcLg.Printf(...) yourself.

// Lots an entrance to a function.
func Enter(format string, v ...interface{}) {
  if TrcLg.enabled {
    TrcLg.Printf("--> " + format, v...)
    indent += "  "
    calldepth += 2
  }
}

// Logs the entrance and immediate exit to a function.
func EnterExit(format string, args ...interface{}) {
  if TrcLg.enabled {
    TrcLg.Printf("- " + format, args...)
  }
}

// Logs the exit a function.
func Exit(format string, args ...interface{}) {
  if TrcLg.enabled {
    indent = indent[2:]
    calldepth -= 2
    TrcLg.Printf("<-- " + format, args...)
  }
}

// Print at or above "trace" verbosity level.
func Trc(fmt string, args ...interface{}) {
  TrcLg.Printf(fmt, args...)
}

// Print at or above "debug" verbosity level.
func Dbg(fmt string, args ...interface{}) {
  DbgLg.Printf(fmt, args...)
}

// Print at or above "verbose" verbosity level.
func Vrb(fmt string, args ...interface{}) {
  VrbLg.Printf(fmt, args...)
}

// Print at or above "info" verbosity level (default).
func Ifo(fmt string, args ...interface{}) {
  IfoLg.Printf(fmt, args...)
}

// Print at or above "warn" verbosity level.
func Wrn(fmt string, args ...interface{}) {
  WrnLg.Printf(fmt, args...)
}

// Print to stderr at or above "error" verbosity level.
func Err(fmt string, args ...interface{}) {
  ErrLg.Printf(fmt, args...)
}

// Print to stderr at or above "fatal" verbosity level.
func Ftl(fmt string, args ...interface{}) {
  FtlLg.Printf(fmt, args...)
}
