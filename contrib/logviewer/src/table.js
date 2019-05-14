import React from 'react';
import PropTypes from 'prop-types';
import classNames from 'classnames';
import { withStyles } from '@material-ui/core/styles';
import TableCell from '@material-ui/core/TableCell';
import TableSortLabel from '@material-ui/core/TableSortLabel';
import Paper from '@material-ui/core/Paper';
import { AutoSizer, Column, SortDirection, Table } from 'react-virtualized';
import Typography from '@material-ui/core/Typography';
import Highlight from 'react-highlight'
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import MuiTable from 'mui-virtualized-table';

const styles = theme => ({
  table: {
    fontFamily: theme.typography.fontFamily,
  },
  flexContainer: {
    display: 'flex',
    alignItems: 'center',
    boxSizing: 'border-box',
  },
  tableRow: {
    cursor: 'pointer',
  },
  tableRowHover: {
    '&:hover': {
      backgroundColor: theme.palette.grey[200],
    },
  },
  tableCell: {
    flex: 1,
  },
  noClick: {
    cursor: 'initial',
  },
});

class MuiVirtualizedTable extends React.PureComponent {
  getRowClassName = ({ index }) => {
    const { classes, rowClassName, onRowClick } = this.props;

    return classNames(classes.tableRow, classes.flexContainer, rowClassName, {
      [classes.tableRowHover]: index !== -1 && onRowClick != null,
    });
  };

  cellRenderer = ({ cellData, columnIndex = null }) => {
    const { columns, classes, rowHeight, onRowClick } = this.props;
    return (
      <TableCell
        component="div"
        className={classNames(classes.tableCell, classes.flexContainer, {
          [classes.noClick]: onRowClick == null,
        })}
        variant="body"
        style={{ height: rowHeight }}
        align={(columnIndex != null && columns[columnIndex].numeric) || false ? 'right' : 'left'}
      >
        {cellData}
      </TableCell>
    );
  };

  headerRenderer = ({ label, columnIndex, dataKey, sortBy, sortDirection }) => {
    const { headerHeight, columns, classes, sort } = this.props;
    const direction = {
      [SortDirection.ASC]: 'asc',
      [SortDirection.DESC]: 'desc',
    };

    const inner = !columns[columnIndex].disableSort && sort != null ? (
        <TableSortLabel active={dataKey === sortBy} direction={direction[sortDirection]}>
          {label}
        </TableSortLabel>
      ) : (
        label
      );

      return (
        <TableCell
          component="div"
          className={classNames(classes.tableCell, classes.flexContainer, classes.noClick)}
          variant="head"
          style={{ height: headerHeight }}
          align={columns[columnIndex].numeric || false ? 'right' : 'left'}
        >
          {inner}
        </TableCell>
      );
  };

  /*
  render() {
    const { classes, columns, ...tableProps } = this.props;
    return (
      <AutoSizer>
        {({ height, width }) => (
          <Table
            className={classes.table}
            height={height}
            width={width}
            {...tableProps}
            rowClassName={this.getRowClassName}
          >
            {columns.map(({ cellContentRenderer = null, className, dataKey, ...other }, index) => {
              let renderer;
              if (cellContentRenderer != null) {
                renderer = cellRendererProps =>
                  this.cellRenderer({
                    cellData: cellContentRenderer(cellRendererProps),
                    columnIndex: index,
                  });
              } else {
                renderer = this.cellRenderer;
              }

              return (
                <Column
                  key={dataKey}
                  headerRenderer={headerProps =>
                    this.headerRenderer({
                      ...headerProps,
                      columnIndex: index,
                    })
                  }
                  className={classNames(classes.flexContainer, className)}
                  cellRenderer={renderer}
                  dataKey={dataKey}
                  {...other}
                />
              );
            })}
          </Table>
        )}
      </AutoSizer>
    );
  }
  */

  render() {
    const { columns } = this.props;

    return <AutoSizer>
       {({ width, height }) => (
         <MuiTable
           data={data}
           columns={[
             {
               name: 'fullName',
               header: 'Name',
               width: 180,
               cell: d => `${d.firstName} ${d.lastName}`,
               cellProps: { style: { paddingRight: 0 } }
             },
             { name: 'jobTitle', header: 'Job Title' },
             { name: 'jobArea', header: 'Job Area' }
           ]}
           width={width}
           includeHeaders={true}
           fixedRowCount={1}
           style={{ backgroundColor: 'transparent' }}
         />
       )}
     </AutoSizer>
  }
}

MuiVirtualizedTable.propTypes = {
  classes: PropTypes.object.isRequired,
  columns: PropTypes.arrayOf(
    PropTypes.shape({
      cellContentRenderer: PropTypes.func,
      dataKey: PropTypes.string.isRequired,
      width: PropTypes.number.isRequired,
    }),
  ).isRequired,
  headerHeight: PropTypes.number,
  onRowClick: PropTypes.func,
  rowClassName: PropTypes.string,
  rowHeight: PropTypes.oneOfType([PropTypes.number, PropTypes.func]),
  sort: PropTypes.func,
};

MuiVirtualizedTable.defaultProps = {
  headerHeight: 56,
  rowHeight: 36,
};

const WrappedVirtualizedTable = withStyles(styles)(MuiVirtualizedTable);

const data = [
  ['Frozen yoghurt', 159, 6.0, 24, 4.0],
  ['Ice cream sandwich', 237, 9.0, 37, 4.3],
  ['Eclair', 262, 16.0, 24, 6.0],
  ['Cupcake', 305, 3.7, 67, 4.3],
  ['Gingerbread', 356, 16.0, 49, 3.9],
];

let id = 0;
function createData(dessert, calories, fat, carbs, protein) {
  id += 1;
  return { id, dessert, calories, fat, carbs, protein };
}

const rows = [];

for (let i = 0; i < 200; i += 1) {
  const randomSelection = data[Math.floor(Math.random() * data.length)];
  rows.push(createData(...randomSelection));
}

class ReactVirtualizedTable extends React.Component {
  constructor (props) {
    super(props)

    console.log(props)
  }

  state = {
    nodes: [],
    records: [],
  }

  setRecord(records, nodes) {
    this.setState({
      records: records,
      nodes: nodes,
    })
  }

  render() {
    var columns = [
      {
        name: 'T',
        header: 'T',
        width: 100,
        cell: d => d.t.elapsed(this.state.records[0].t),
        cellProps: { style: { paddingRight: 0 } }
      },
    ]

    var width = (100 / this.state.nodes.length)

    this.state.nodes.map(node => {
      columns.push({
        name: node,
        header: node,
        //width: 100,
        cell: d => {
          if (d.node != node) {
            return ''
          }

          return <RecordItem record={d} classes={{}}/>
        }
      })
    })

    return (
      <Paper style={{ height: '100%', width: '100%' }}>
        <AutoSizer>
           {({ width, height }) => (
             <MuiTable
               data={this.state.records}
               columns={columns}
               width={width}
               includeHeaders={true}
               fixedRowCount={1}
               style={{ backgroundColor: 'transparent' }}
             />
           )}
         </AutoSizer>
      </Paper>
    );
  }
}


class RecordItem extends React.Component {
  render() {
    const { classes } = this.props;
    const { record } = this.props;

    return (
          <ExpansionPanel>
            <ExpansionPanelSummary
                expandIcon={<ExpandMoreIcon />}>
              <div className={classes.column}>
                <Typography className={classes.heading}>{record.message}</Typography>
              </div>
            </ExpansionPanelSummary>
            <ExpansionPanelDetails className={classes.details}>
                <Highlight className='json'>{JSON.stringify(record.extra, null, 2)}</Highlight>
            </ExpansionPanelDetails>
          </ExpansionPanel>
    )
  }
}



export default ReactVirtualizedTable;
