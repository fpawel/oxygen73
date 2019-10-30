
include "apitypes.thrift"

service MainSvc {
   apitypes.Party getParty(1:i64 partyID)
   list<apitypes.Product> listProducts(1:i64 partyID)
   list<apitypes.YearMonth> listYearMonths()
   list<apitypes.Bucket> listBucketsOfYearMonth(1:i32 year, 2:i32 month)

   oneway void requestMeasurements(1:apitypes.TimeUnixMillis timeFrom, 2:apitypes.TimeUnixMillis timeTo)
   oneway void requestProductMeasurements(1:i64 bucketID, 2:i32 place)

   oneway void createNewParty()

   list<apitypes.Product> listLastPartyProducts()
   void setProductSerialAtPlace(1:i32 place, 2:i32 serial)
   void deleteProductAtPlace(1:i32 place)

   list<apitypes.ProductBucket> findProductsBySerial(1:i32 serial)

   list<apitypes.TimeUnixMillis> ListLogEntriesDays()
   list<apitypes.LogEntry> LogEntriesOfDay(1:apitypes.TimeUnixMillis daytime, 2:string filter)

   apitypes.AppConfig getAppConfig()
   oneway void setAppConfig(1:apitypes.AppConfig appConfig)

   string getAppConfigToml()
   oneway void setAppConfigToml(1:string appConfigToml)
}