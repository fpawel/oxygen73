include "apitypes.thrift"

service GuiSvc {
   oneway void notifyWriteConsole(1:string str)
   oneway void notifyStatus(1:bool ok, 2:string str)
   oneway void notifyMeasurement(1:apitypes.Measurement measurement)
}