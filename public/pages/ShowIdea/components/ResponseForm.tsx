import * as React from "react";

import { Modal, Button, DisplayError, Textarea } from "@fider/components/common";
import { Comment, Idea, IdeaStatus, User } from "@fider/models";
import { IdeaSearch } from "../";

import { actions, Failure } from "@fider/services";

interface ResponseFormProps {
  idea: Idea;
}

interface ResponseFormState {
  showModal: boolean;
  status: number;
  text: string;
  originalNumber: number;
  error?: Failure;
}

export class ResponseForm extends React.Component<ResponseFormProps, ResponseFormState> {
  private modal!: HTMLDivElement;

  constructor(props: ResponseFormProps) {
    super(props);

    this.state = {
      showModal: false,
      status: this.props.idea.status,
      originalNumber: 0,
      text: this.props.idea.response && this.props.idea.response.text
    };
  }

  private async submit() {
    const result = await actions.respond(this.props.idea.number, this.state);
    if (result.ok) {
      location.reload();
    } else {
      this.setState({
        error: result.error
      });
    }
  }

  public render() {
    const button = (
      <Button className="respond" size="small" fluid={true} onClick={async () => this.setState({ showModal: true })}>
        <i className="announcement icon" /> Respond
      </Button>
    );

    const modal = (
      <Modal.Window isOpen={this.state.showModal} center={false} size="large">
        <Modal.Content>
          <div className="ui form fdr-response-form">
            <DisplayError fields={["status"]} error={this.state.error} />
            <div className="two fields">
              <div className="field">
                <label>Status</label>
                <select
                  className="ui dropdown"
                  defaultValue={this.props.idea.status.toString()}
                  onChange={e =>
                    this.setState({
                      status: parseInt(e.currentTarget.value, 10)
                    })
                  }
                >
                  {IdeaStatus.All.map(s => (
                    <option key={s.value} value={s.value.toString()}>
                      {s.title}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            {this.state.status === IdeaStatus.Duplicate.value ? (
              <>
                <DisplayError fields={["originalNumber"]} error={this.state.error} />
                <IdeaSearch
                  exclude={[this.props.idea.number]}
                  onChanged={originalNumber => this.setState({ originalNumber })}
                />
                <span className="info">Votes from this idea will be merged into original idea.</span>
              </>
            ) : (
              <>
                <DisplayError fields={["text"]} error={this.state.error} />
                <div className="field">
                  <Textarea
                    onChange={e => this.setState({ text: e.currentTarget.value })}
                    defaultValue={this.state.text}
                    placeholder="What's going on with this idea? Let your users know what are your plans..."
                  />
                </div>
              </>
            )}
          </div>
        </Modal.Content>

        <Modal.Footer>
          <Button color="green" onClick={() => this.submit()}>
            Submit
          </Button>
          <Button onClick={async () => this.setState({ showModal: false })}>Cancel</Button>
        </Modal.Footer>
      </Modal.Window>
    );

    return (
      <>
        {button}
        {modal}
      </>
    );
  }
}
