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

	const withdrawRequest = async () => {
		let input = getFormData();
		console.log(input);
		const data = new URLSearchParams();
		data.append = ("to_address", input.to_address);
		data.append = ("amount", input.amount);

		fetch(`${URL}/withdraw/`, {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
			},
			body: data,
		})
			.then((res) => res.json())
			.then((data) => console.log(data))
			.catch((error) => console.error(error));
	};

	const statusBar = document.createElement("p");

	const getFormData = () => {
		const form = document.querySelector("form");
		return (data = Object.fromEntries(new FormData(form).entries()));
	};
	$("#withdraw-btn").click(function () {
		console.log("withdraw click...");
		withdrawRequest();
	});

	getStatus();
});
