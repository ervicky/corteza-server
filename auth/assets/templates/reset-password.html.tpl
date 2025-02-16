{{ template "inc_header.html.tpl" . }}
<div class="card-body p-0">
	<h4 class="card-title p-3 border-bottom">Reset your password</h4>
	<form
		method="POST"
		action="{{ links.ResetPassword }}"
		class="p-3"
	>
		{{ .csrfField }}
		{{ if .form.error }}
		<div class="text-danger font-weight-bold p-3" role="alert">
			{{ .form.error }}
		</div>
		{{ end }}
		<div class="mb-3">
			<label>
                E-mail *
            </label>
			<input
				type="email"
				class="form-control"
				name="email"
				readonly
				placeholder="email@domain.ltd"
				value="{{ .user.Email }}"
				aria-label="Email">
		</div>
		<div class="mb-3">
            <label>
                New Password *
            </label>
			<input
				type="password"
				required
				class="form-control"
				name="password"
				autocomplete="new-password"
				placeholder="Set new password"
				aria-label="Set new password">
		</div>
		<div class="text-right">
			<button class="btn btn-primary btn-block btn-lg" type="submit">Change your password</button>
		</div>
	</form>
</div>
{{ template "inc_footer.html.tpl" . }}
