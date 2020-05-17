import React from 'react';

class Numbers extends React.Component {

    constructor(props) {
    super(props);

    this.state = {
      animate: false,
    };
    }

componentDidUpdate(prevProps) {
  // Typical usage (don't forget to compare props):
    if ((this.props.count !== prevProps.count) && !this.state.animate && this.props.count > 1) {
    this.setState({animate: true});
  } else {
      if (this.state.animate) {
   this.timerID = setTimeout(() => {
      this.setState({animate: false});
    }, 1050);
      }
  }
  
}
componentWillUnmount() {
    clearInterval(this.timerID);
  }
    render() {
        return (
            <div className={`numbers ${this.state.animate ? "shake-animation count-text-static" : ""}`}>
                <div className="flip-card-inner">
                <div className="numbers-front">{this.props.numbers}</div>
                <div className="count-back">{this.props.count}</div>
                </div>
            </div>
        )
}
}

export default Numbers;
