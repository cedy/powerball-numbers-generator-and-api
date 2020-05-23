import React from 'react';

class HistoryRow extends React.Component {
    
    constructor(props) {
        super(props);

        this.state = {
            animate:false,
        };
    }

    render() {
        return (
        // date, history number, randomly generated number, count
            <tr>
                <td className="text-nowrap">{this.props.date}</td>
                <td className="text-nowrap">{this.props.numbers}</td>
                <td>{this.props.dayCount}</td>
                <td>{this.props.weekCount}</td>
                <td>{this.props.monthCount}</td>
                <td>{this.props.yearCount}</td>
                <td>{this.props.allTime}</td>
            </tr>
        )
    }
}

export default HistoryRow
