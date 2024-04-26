const ctx1 = document.querySelector("#myChart").getContext("2d");
var myChart = new Chart(ctx1, {
	type: "line",
	plugins: [ChartDatasourcePrometheusPlugin],
	options: {
		plugins: {
			"datasource-prometheus": {
				prometheus: {
					endpoint: "http://localhost:9092",
				},
				query:
					"sequoia_current_proofs_processing or sequoia_file_count offset 10m",
				timeRange: {
					type: "relative",
					// from 24 hours ago to now
					start: -24 * 60 * 60 * 1000,
					end: 0,
				},
			},
		},
	},
});

const ctx2 = document.querySelector("#networkChart").getContext("2d");
var networkChart = new Chart(ctx2, {
	type: "line",
	plugins: [ChartDatasourcePrometheusPlugin],
	options: {
		plugins: {
			"datasource-prometheus": {
				prometheus: {
					endpoint: "http://localhost:9092",
				},
				query: "sequoia_block_height",
				timeRange: {
					type: "relative",
					// from 2 hours ago to now
					start: -2 * 60 * 60 * 1000,
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
