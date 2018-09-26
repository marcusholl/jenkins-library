def call(Map parameters = [:], body) {

    def cpe = parameters.cpe
    def measurementName = parameters.get('measurementName', 'test_duration')

    //start measurement
    def start = System.currentTimeMillis()

    body()

    //record measurement
    def duration = System.currentTimeMillis() - start

    if (cpe != null)
        cpe.setPipelineMeasurement(measurementName, duration)

    return duration
}

