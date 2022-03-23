import { h } from "preact";
import classes from "./index.module.scss";
import { Icon, Button, Grid } from "semantic-ui-react";

export default function NavBar() {
	return (
		<div class={classes["navbar-container"]}>
			<Grid>
				<Grid.Column floated="left" width={1}>
					<Button icon>
						<Icon name="world" />
					</Button>

					<Button icon>
						<Icon name="world" />
					</Button>
				</Grid.Column>
			</Grid>
		</div>
	);
}
