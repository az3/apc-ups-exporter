package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Create the metrics
var (

	// Status (as number) - STATUS
	metricStatus = promauto.NewGaugeVec( prometheus.GaugeOpts {
		Namespace: "ups",
		Name: "status",
		Help: "The current status. STATUS The current status of the UPS (1-ONLINE, 2-ONBATT, etc.).",
	}, []string{"string"})

	/*************************************/

	// Expected power input (as voltage) - NOMINV
	metricPowerInputExpectVoltage = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "power",
		Name: "input_expect_voltage",
		Help: "The expected input voltage. NOMINV The input voltage that the UPS is configured to expect.",
	} )

	// Maximum power output (as wattage) - NOMPOWER
	metricPowerOutputWattage = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "power",
		Name: "output_maximum_wattage",
		Help: "The maximum power the UPS can output. NOMPOWER The maximum power in Watts that the UPS is designed to supply.",
	} )

	// Current line voltage (as voltage) - LINEV
	metricPowerLineVoltage = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "power",
		Name: "line_voltage",
		Help: "The current line voltage as returned by the UPS. LINEV The current line voltage as returned by the UPS.",
	} )

	// Current load capacity (as percentage) - LOADPCT
	metricPowerLoadPercent = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "power",
		Name: "load_percent",
		Help: "The current load capacity as estimated by the UPS, as a percentage. LOADPCT The percentage of load capacity as estimated by the UPS.",
	} )

	/*************************************/

	// Expected power output of the battery (as voltage) - NOMBATTV
	metricBatteryExpectVoltage = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "battery",
		Name: "output_expect_voltage",
		Help: "The expected output voltage of the battery. NOMBATTV The nominal battery voltage.",
	} )

	// Actual power output of the battery (as voltage) - BATTV
	metricBatteryActualVoltage = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "battery",
		Name: "output_actual_voltage",
		Help: "The actual output voltage of the battery. BATTV Battery voltage as supplied by the UPS.",
	} )

	// Latest time spent on battery (in seconds) - TONBATT
	metricBatteryTimeSpentLatestSeconds = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "battery",
		Name: "time_spent_latest_seconds",
		Help: "The latest time spent on battery. TONBATT Time in seconds currently on batteries, or 0.",
	} )

	// Total time spent on battery (in seconds) - CUMONBATT
	metricBatteryTimeSpentTotalSeconds = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "battery",
		Name: "time_spent_total_seconds",
		Help: "The total time spent on battery. CUMONBATT Total (cumulative) time on batteries in seconds since apcupsd startup.",
	} )

	// Remaining charge of the battery (as percentage) - BCHARGE
	metricBatteryRemainingChargePercent = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "battery",
		Name: "remaining_charge_percent",
		Help: "The remaining charge on the battery, as a percentage. BCHARGE The percentage charge on the batteries.",
	} )

	// Remaining time of the battery (in minutes) - TIMELEFT
	metricBatteryRemainingTimeMinutes = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "battery",
		Name: "remaining_time_minutes",
		Help: "The remaining runtime left on the battery as estimated by the UPS, in minutes. TIMELEFT The remaining runtime left on batteries as estimated by the UPS.",
	} )

	/*************************************/

	// Configured minimum battery charge (as percentage) - MBATTCHG
	metricDaemonRemainingChargePercent = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "remaining_charge_percent",
		Help: "The configured minimum remaining charge on the battery to trigger a system shutdown, as a percentage. MBATTCHG If the battery charge percentage (BCHARGE) drops below this value, apcupsd will shutdown your system. Value is set in the configuration file (BATTERYLEVEL)",
	} )

	// Configured minimum battery remaining time (in minutes) - MINTIMEL
	metricDaemonRemainingTimeMinutes = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "remaining_time_minutes",
		Help: "The configured minimum remaining runtime left on the battery to trigger a system shutdown, in minutes. MINTIMEL apcupsd will shutdown your system if the remaining runtime equals or is below this point. Value is set in the configuration file (MINUTES)",
	} )

	// Configured maximum timeout (in minutes) - MAXTIME
	metricDaemonTimeoutMinutes = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "timeout_minutes",
		Help: "The configured maximum time running on the battery to trigger a system shutdown, in minutes. MAXTIME apcupsd will shutdown your system if the time on batteries exceeds this value. A value of zero disables the feature. Value is set in the configuration file (TIMEOUT)",
	} )

	// Number of transfers to battery - NUMXFERS
	metricDaemonTransferCount = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "transfer_count",
		Help: "The number of transfers to the battery. NUMXFERS The number of transfers to batteries since apcupsd startup.",
	} )

	// Daemon startup time (as unix timestamp) - STARTTIME
	metricDaemonStartTimestamp = promauto.NewGauge( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "start_timestamp",
		Help: "The date & time the daemon was started. STARTTIME The time/date that apcupsd was started.",
	} )

	// Daemon HOSTNAME
	metricHostname = promauto.NewGaugeVec( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "hostname",
		Help: "HOSTNAME The name of the machine that collected the UPS data.",
	}, []string{"string"})

	metricApcHeader = promauto.NewGaugeVec( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "apc",
		Help: "APC Header record indicating the STATUS format revision level, the number of records that follow the APC statement, and the number of bytes that follow the record.",
	}, []string{"string"})

	metricAlarmDel = promauto.NewGaugeVec( prometheus.GaugeOpts {
		Namespace: "ups",
		Subsystem: "daemon",
		Name: "alarm",
		Help: "ALARMDEL The delay period for the UPS alarm.",
	}, []string{"string"})
)

// Sets all of the metrics to zero
func ResetMetrics() {

	// Status
	metricStatus.WithLabelValues("UNKNOWN0").Set( 0 )

	// Power
	metricPowerInputExpectVoltage.Set( 0 )
	metricPowerOutputWattage.Set( 0 )
	metricPowerLineVoltage.Set( 0 )
	metricPowerLoadPercent.Set( 0 )

	// Battery
	metricBatteryExpectVoltage.Set( 0 )
	metricBatteryActualVoltage.Set( 0 )
	metricBatteryTimeSpentLatestSeconds.Set( 0 )
	metricBatteryTimeSpentTotalSeconds.Set( 0 )
	metricBatteryRemainingChargePercent.Set( 0 )
	metricBatteryRemainingTimeMinutes.Set( 0 )

	// Daemon
	metricDaemonRemainingChargePercent.Set( 0 )
	metricDaemonRemainingTimeMinutes.Set( 0 )
	metricDaemonTimeoutMinutes.Set( 0 )
	metricDaemonTransferCount.Set( 0 )
	metricDaemonStartTimestamp.Set( 0 )

	metricHostname.WithLabelValues("localhost").Set(0)
	metricApcHeader.WithLabelValues("001,037,0906").Set(0)
	// metricAlarmDel.WithLabelValues("No alarm").Set(-1)

}

// Serves the metrics page over HTTP
func ServeMetrics( address net.IP, port int, path string ) ( err error ) {

	// Handle requests to the metrics path using the Prometheus HTTP handler
	http.Handle( path, promhttp.Handler() )

	// Listen for HTTP requests
	listenError := http.ListenAndServe( fmt.Sprintf( "%s:%d" , address, port ), nil )
	if listenError != nil { return listenError }

	// No error, all was good
	return nil

}

/*
-- current values in /metrics

BATTV
BCHARGE
CUMONBATT
HOSTNAME
LINEV
LOADPCT
MAXTIME
MBATTCHG
MINTIMEL
NOMBATTV
NOMINV
NOMPOWER
NUMXFERS
STARTTIME
STATUS
TIMELEFT
TONBATT

-- all values supplied by apcupsd

ALARMDEL
APC
BATTDATE
BATTV
BCHARGE
CABLE
CUMONBATT
DATE
DRIVER
ENDAPC
HITRANS
HOSTNAME
LASTSTEST
LASTXFER
LINEV
LOADPCT
LOTRANS
MAXTIME
MBATTCHG
MINTIMEL
MODEL
NOMBATTV
NOMINV
NOMPOWER
NUMXFERS
SELFTEST
SENSE
SERIALNO
STARTTIME
STATFLAG
STATUS
TIMELEFT
TONBATT
UPSMODE
UPSNAME
VERSION
XOFFBATT
XONBATT

-- daemon values:
APC      : 001,037,0937
DATE     : 2024-03-12 21:32:47 +0300
HOSTNAME : AzurPc
VERSION  : 3.14.14 (31 May 2016) mingw
UPSNAME  : AzurPc
CABLE    : USB Cable
DRIVER   : USB UPS Driver
UPSMODE  : Stand Alone
STARTTIME: 2024-03-12 13:41:28 +0300
MODEL    : Back-UPS BX2200MI
STATUS   : ONLINE
LINEV    : 232.0 Volts
LOADPCT  : 14.0 Percent
BCHARGE  : 99.0 Percent
TIMELEFT : 22.6 Minutes
MBATTCHG : 5 Percent
MINTIMEL : 3 Minutes
MAXTIME  : 0 Seconds
SENSE    : Medium
LOTRANS  : 145.0 Volts
HITRANS  : 295.0 Volts
ALARMDEL : No alarm
BATTV    : 27.2 Volts
LASTXFER : Automatic or explicit self test
NUMXFERS : 2
XONBATT  : 2024-03-12 15:04:48 +0300
TONBATT  : 0 Seconds
CUMONBATT: 14 Seconds
XOFFBATT : 2024-03-12 15:04:55 +0300
LASTSTEST: 2024-03-12 15:04:48 +0300
SELFTEST : OK
STATFLAG : 0x05000008
SERIALNO : 9B2334A38053
BATTDATE : 2024-03-12
NOMINV   : 230 Volts
NOMBATTV : 24.0 Volts
NOMPOWER : 1200 Watts
END APC  : 2024-03-12 21:33:32 +0300

*/