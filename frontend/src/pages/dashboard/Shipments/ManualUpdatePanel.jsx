import React from 'react';

export default function ManualUpdatePanel({ onSubmit, submittingUpdate, rawUpdate, setRawUpdate, updateMsg }) {
	return (
		<div className="glass-card panel-card manual-update-card">
			<h3>Simulate Carrier Update</h3>
			<p className="panel-desc">Simulate a raw email or tracking update from a carrier to trigger the Operations Agent pipeline.</p>

			<form onSubmit={onSubmit}>
				<textarea
					className="raw-update-textarea"
					rows={3}
					placeholder="e.g., Vessel MAERSK TITAN delayed 3 days due to typhoon in the Pacific. Revised ETA Hamburg is now 29 August."
					value={rawUpdate}
					onChange={(e) => setRawUpdate(e.target.value)}
				/>
				<button
					type="submit"
					className="btn-primary btn-submit-update"
					disabled={submittingUpdate || !rawUpdate.trim()}
				>
					{submittingUpdate ? 'Processing Event...' : 'Submit Tracking Update'}
				</button>
			</form>
			{updateMsg && <p className="update-msg">{updateMsg}</p>}
		</div>
	);
}
