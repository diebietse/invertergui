package munin

const muninConfig = `multigraph in_batvolt
graph_title Battery Voltage
graph_vlabel Voltage (V)
graph_category inverter
graph_info Battery voltage

volt.info Voltage of battery
volt.label Voltage of battery (V)

multigraph in_batcharge
graph_title Battery Charge
graph_vlabel Charge (%)
graph_category inverter
graph_info Battery charge

charge.info Estimated charge of battery
charge.label Battery charge (%)

multigraph in_batcurrent
graph_title Battery Current
graph_vlabel Current (A)
graph_category inverter
graph_info Battery current

current.info Battery current
current.label Battery current (A)

multigraph in_batpower
graph_title Battery Power
graph_vlabel Power (W)
graph_category inverter
graph_info Battery power

power.info Battery power
power.label Battery power (W)

multigraph in_mainscurrent
graph_title Mains Current
graph_vlabel Current (A)
graph_category inverter
graph_info Mains current

currentin.info Input current
currentin.label Input current (A)
currentout.info Output current
currentout.label Output current (A)

multigraph in_mainsvoltage
graph_title Mains Voltage
graph_vlabel Voltage (V)
graph_category inverter
graph_info Mains voltage

voltagein.info Input voltage
voltagein.label Input voltage (V)
voltageout.info Output voltage
voltageout.label Output voltage (V)

multigraph in_mainspower
graph_title Mains Power
graph_vlabel Power (VA)
graph_category inverter
graph_info Mains power

powerin.info Input power
powerin.label Input power (VA)
powerout.info Output power
powerout.label Output power (VA)

multigraph in_mainsfreq
graph_title Mains frequency
graph_vlabel Frequency (Hz)
graph_category inverter
graph_info Mains frequency

freqin.info In frequency
freqin.label In frequency (Hz)
freqout.info Out frequency
freqout.label Out frequency (Hz)
`
