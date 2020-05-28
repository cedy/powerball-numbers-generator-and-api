import React from 'react';
import Pagination from 'react-bootstrap/Pagination'


class PPagination extends React.Component {
    
    render() {
        return (
        <Pagination size="lg">
          <Pagination.First onClick={this.props.first}/>
          <Pagination.Prev onClick={this.props.previous}/>
          <Pagination.Item active>{this.props.currentPage}</Pagination.Item>
          <Pagination.Next onClick={this.props.next}/>
        </Pagination>)
    }
}

export default PPagination;
