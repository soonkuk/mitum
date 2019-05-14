import React from 'react';
import PropTypes from 'prop-types';
import { withStyles } from '@material-ui/core/styles';
import Dialog from '@material-ui/core/Dialog';
import blue from '@material-ui/core/colors/blue';
import Files from './file';

const styles = {
  avatar: {
    backgroundColor: blue[100],
    color: blue[600],
  },
};

class SimpleDialog extends React.Component {
  render() {
    const { classes, onClose, ...other } = this.props;

    console.log('aaaaaaaa', other)
    return (
      <Dialog aria-labelledby="simple-dialog-title" {...other}>
        <Files onSelected={other.onDrop} />
      </Dialog>
    );
  }
}

SimpleDialog.propTypes = {
  classes: PropTypes.object.isRequired,
  onClose: PropTypes.func,
  onBlur: PropTypes.func,
  onSelected: PropTypes.func,
};

export default withStyles(styles)(SimpleDialog);
