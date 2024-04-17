const ctx = document.querySelector("#myChart").getContext("2d");
var myChart = new Chart(ctx, {
	type: "line",
	plugins: [ChartDatasourcePrometheusPlugin],
	options: {
		plugins: {
			"datasource-prometheus": {
				prometheus: {
					endpoint: "https://prometheus.demo.do.prometheus.io",
					baseURL: "/api/v1", // default value
				},
				query: "sum by (job) (go_gc_duration_seconds)",
				timeRange: {
					type: "relative",

					// from 12 hours ago to now
					start: -12 * 60 * 60 * 1000,
					end: 0,
				},
			},
		},
	},
});

const doSomething = () => {
	console.log("hello...");
};

window.addEventListener("onload", doSomething(), false);
