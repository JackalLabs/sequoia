$(document).ready(function () {
	const URL = "http://localhost:3334";

	const getStatus = async () => {
		let data = await fetch(URL).then((res) => res.json());
		let version = await fetch(`${URL}/version`).then((res) => res.json());

		if (data) {
			let status = `<p>Status: ${data.status}</p>`;
			let address = `<p>Wallet: ${data.address}</p>`;
			$("#status-bar").append(status, address);
		}
	};

	const statusBar = document.createElement("p");

	$("#withdraw-btn").click(function () {
		console.log("withdraw click");
	});

	getStatus();
});
