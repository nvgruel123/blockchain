import React, { Component } from 'react';
import "./Main.css"

class Main extends Component {

  constructor(props) {

    super(props)

    this.state = {
      data: null,
    }
  }

  handlesubmit = (event) => {
    event.preventDefault()
    var height = event.target.height.value
    var hash = event.target.hash.value
    var formData = "height=" + height + "&hash=" + hash
    fetch("http://localhost:3053/block/display", {
      method: "POST",
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
      },
      body: formData
    })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      this.setState({
        data: json.mpt
      })
    })
    .catch((err) => {
      console.log(err);
    })  
  }

  renderData = () => {
    if (this.state.data) {
      var time = new Date(this.state.data.time_caught*1000).toString()
      console.log(time)
      return (
        <div>
          <p>Count: {this.state.data.count}</p>
          <p>Time Caught: {time}</p>
          <p>Type: {this.state.data.type}</p>
          <p>Weight: {this.state.data.weight} lbs</p>

        </div>
      )
    }
  }

  render() {
    return (
      <div className="display">
        <form onSubmit={this.handlesubmit}>
          <p><input type="text" name="height" id="height" placeholder="Block Height" onChange={this.handleChange}/></p>
          <p><input type="text" name="hash" id="hash" placeholder="Block Hash" onChange={this.handleChange}/></p>
          <button type="submit">check</button>
          {this.renderData()}
        </form>
        <div>
        </div>
      </div>
    );
  }
}

export default Main;
